package controller

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/utils"
	informersastrocorev1 "AstroKube/pkg/client/informers/externalversions/core/v1"
	"context"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type ClusterController struct {
	ClusterName        string
	Cluster            *astrov1.Cluster
	SubExternalIp      string
	NamespacePrefix    string
	HeartbeatFrequency time.Duration

	SubApiServerProtocol string
	SubApiServerPort     int32

	queue workqueue.RateLimitingInterface

	clusterDaemonEndpoint astrov1.ClusterDaemonEndpoints
	apiServerAddress      []astrov1.Address
	clusterInfo           astrov1.ClusterInfo
	secretRef             astrov1.ClusterSecretRef

	clusterInformer informersastrocorev1.ClusterInformer
}

func NewClusterController(
	clusterName string,
	cluster *astrov1.Cluster,
	subExternalIp string,
	heartbeatFrequency time.Duration,
	namespacePrefix string,
	subApiServerProtocol string,
	subApiServerPort int32,
	informer informersastrocorev1.ClusterInformer,
) *ClusterController {
	cc := ClusterController{
		ClusterName:          clusterName,
		Cluster:              cluster,
		SubExternalIp:        subExternalIp,
		HeartbeatFrequency:   heartbeatFrequency,
		clusterInformer:      informer,
		NamespacePrefix:      namespacePrefix,
		SubApiServerPort:     subApiServerPort,
		SubApiServerProtocol: subApiServerProtocol,
	}

	cc.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster-controller-pod-queue")
	cache.SubSharedInformerFactory.Core().V1().Pods().Informer().AddEventHandler(
		toolscache.ResourceEventHandlerFuncs{
			AddFunc:    cc.addPodEvent,
			UpdateFunc: cc.upadtePodEvent,
			DeleteFunc: cc.deletePodEvent,
		})

	cache.SubSharedInformerFactory.Core().V1().Nodes().Informer().AddEventHandler(
		toolscache.ResourceEventHandlerFuncs{
			AddFunc:    cc.addNodeEvent,
			UpdateFunc: cc.updateNodeEvent,
			DeleteFunc: cc.deleteNodeEvent,
		})

	if cc.clusterInformer != nil {
		cc.clusterInformer.Informer().AddEventHandler(toolscache.ResourceEventHandlerFuncs{
			AddFunc:    cc.addClusterEvent,
			UpdateFunc: cc.updateClusterEvent,
			DeleteFunc: cc.deleteClusterEvent,
		})
	}

	return &cc
}

type event struct {
	eventType string
	name      string
}

func (cc *ClusterController) addPodEvent(obj interface{}) {
	pod := obj.(*apicorev1.Pod)
	// 生成唯一标识符
	key, err := toolscache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		klog.Errorf("addPodEvent error: %s", err)
		return
	}
	klog.V(4).Infof("add pod: %s", key)
	cc.queue.Add(&event{eventType: "Pod", name: key})
}

func (cc *ClusterController) upadtePodEvent(old, new interface{}) {
	oldPod := old.(*apicorev1.Pod)
	newPod := old.(*apicorev1.Pod)
	if oldPod.Spec.NodeName == newPod.Spec.NodeName && oldPod.Status.Phase == newPod.Status.Phase {
		return
	}
	key, err := toolscache.MetaNamespaceKeyFunc(newPod)
	if err != nil {
		klog.Errorf("updatePodEvent error: %s", err)
		return
	}
	klog.V(4).Infof("update pod: %s", key)
	cc.queue.Add(&event{
		eventType: "Pod",
		name:      key,
	})
}

func (cc *ClusterController) deletePodEvent(obj interface{}) {
	pod := obj.(*apicorev1.Pod)
	key, err := toolscache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		klog.Errorf("deletePodEvent error: %s", err)
		return
	}
	klog.V(4).Infof("delete pod: %s", key)
	cc.queue.Add(&event{eventType: "Pod", name: key})
}

func (cc *ClusterController) addNodeEvent(obj interface{}) {
	node := obj.(*apicorev1.Node)
	klog.V(4).Infof("add node: %s", node)
	cc.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (cc *ClusterController) updateNodeEvent(old, new interface{}) {
	node := new.(*apicorev1.Node)
	klog.V(4).Infof("update node: %s", node.Name)
	cc.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (cc *ClusterController) deleteNodeEvent(obj interface{}) {
	node := obj.(*apicorev1.Node)
	klog.V(4).Infof("add node: %s", node)
	cc.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (cc *ClusterController) addClusterEvent(obj interface{}) {
	cluster := obj.(*astrov1.Cluster)
	klog.V(4).Infof("add cluster: %s", cluster.Name)
	cc.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}

func (cc *ClusterController) updateClusterEvent(old, new interface{}) {
	cluster := new.(*astrov1.Cluster)
	klog.V(4).Infof("update cluster: %s", cluster.Name)
	cc.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}

func (cc *ClusterController) deleteClusterEvent(obj interface{}) {
	cluster := obj.(*astrov1.Cluster)
	klog.V(4).Infof("delete cluster: %s", cluster.Name)
	cc.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}

func (cc *ClusterController) Run(ctx context.Context) {
	utils.ContextStart(ctx)
	defer utils.ContextStop(ctx)
	klog.Info("starting astro cluster controller")

	go wait.UntilWithContext(ctx, cc.updateStatus, 5*time.Minute)

}

func (cc *ClusterController) updateStatus(ctx context.Context) {

}
