package controller

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	clientcache "AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/utils"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

const (
	Ready                 = "Ready"
	NoTReady              = "NotReady"
	StatusUpdated         = "StatusUpdated"
	ApiServerIsUp         = "ApiServerIsUp"
	SchedulerIsUp         = "SchedulerIsUp"
	ControllerManagerIsUp = "ControllerManagerIsUp"
	EtcdIsUp              = "EtcdIsUp"
)

type ComponentsController struct {
	RenewInterval time.Duration
	Client        kubernetes.Interface
	ClusterName   string
}

func NewComponentsController(client kubernetes.Interface, clusterName string, renewInterval time.Duration) *ComponentsController {
	return &ComponentsController{
		RenewInterval: renewInterval,
		Client:        client,
		ClusterName:   clusterName,
	}
}

func (c ComponentsController) Run(ctx context.Context) {
	utils.ContextStart(ctx)
	defer utils.ContextStop(ctx)
	wait.UntilWithContext(ctx, c.sync, c.RenewInterval)
}

func (c ComponentsController) sync(ctx context.Context) {
	cluster, err := clientcache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
	if err != nil {
		klog.Error("cluster get failed from Core cluster: ", err.Error())
		return
	}
	nCluster := cluster.DeepCopy()
	utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
		Type:    StatusUpdated,
		Status:  astrov1.TrueCondition,
		Reason:  "status updated",
		Message: "status updated",
	})
	componentStatuses, err := c.Client.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
			Type:    ApiServerIsUp,
			Status:  astrov1.FalseCondition,
			Message: err.Error(),
			Reason:  err.Error(),
		})
		utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
			Type:    Ready,
			Status:  astrov1.FalseCondition,
			Message: fmt.Sprintf("apiserver is down: %s", err.Error()),
			Reason:  fmt.Sprintf("apiserver is down: %s", err.Error()),
		})
		_, err = clientcache.CoreClientCache.AstroClient.AstroV1().Clusters("default").UpdateStatus(context.TODO(), nCluster, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		}
		return
	}
	utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
		Type:    ApiServerIsUp,
		Status:  astrov1.TrueCondition,
		Message: "OK",
		Reason:  "OK",
	})

	ready := true
	var readyReason string
	typeMap := map[string]string{
		"controller-manager": ControllerManagerIsUp,
		"scheduler":          SchedulerIsUp,
	}
	etcdHealth := true
	var etcdHealthReason string
	for _, componentStatus := range componentStatuses.Items {
		for _, cond := range componentStatus.Conditions {
			if cond.Type == "Healthy" {
				if t, ok := typeMap[componentStatus.Name]; ok {
					utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
						Type:    t,
						Status:  astrov1.ConditionStatus(cond.Status),
						Reason:  cond.Message,
						Message: cond.Message,
					})
					if cond.Status != v1.ConditionTrue {
						ready = false
						readyReason = fmt.Sprintf("component %s health false: %s", componentStatus.Name, cond.Message)
					}
				} else if strings.HasPrefix(componentStatus.Name, "etcd") {
					if cond.Status != v1.ConditionTrue {
						etcdHealth = false
						etcdHealthReason = fmt.Sprintf("component %s health false: %s", componentStatus.Name, cond.Message)
						ready = false
						readyReason = fmt.Sprintf("componet %s health false: %s", componentStatus.Name, cond.Message)
					}
				}
			}
		}
	}

	var etcdHealthStatus string
	if etcdHealth {
		etcdHealthStatus = astrov1.TrueCondition
		readyReason = "OK"
	} else {
		etcdHealthStatus = astrov1.FalseCondition
	}
	utils.UpdateClusterCondition(nCluster, &astrov1.ClusterCondition{
		Type:    EtcdIsUp,
		Status:  astrov1.ConditionStatus(etcdHealthStatus),
		Reason:  etcdHealthReason,
		Message: etcdHealthReason,
	})

	var clusterCond astrov1.ClusterCondition
	if ready {
		clusterCond = astrov1.ClusterCondition{Type: Ready, Status: astrov1.TrueCondition, Reason: readyReason, Message: "astrolet ready now"}
	} else {
		clusterCond = astrov1.ClusterCondition{Type: NoTReady, Status: astrov1.FalseCondition, Reason: readyReason, Message: "astrolet not ready now"}
	}
	utils.UpdateClusterCondition(nCluster, &clusterCond)
	_, err = clientcache.CoreClientCache.AstroClient.AstroV1().Clusters("default").UpdateStatus(ctx, nCluster, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
	} else {
		klog.V(4).Infof("cluster component condition update successfully")
	}
}

func (c *ComponentsController) newCluster(ctx context.Context) *astrov1.Cluster {
	return &astrov1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ClusterName,
			Labels: map[string]string{
				astrov1.ClusterLevelLabel:    "1",
				astrov1.ClusterLabel:         c.ClusterName,
				astrov1.ClusterFullNameLabel: c.ClusterName,
			},
		},
		Spec: astrov1.ClusterSpec{},
	}
}
