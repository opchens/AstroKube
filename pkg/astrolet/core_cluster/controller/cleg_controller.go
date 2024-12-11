package controller

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/utils"
	informersastrov1 "AstroKube/pkg/client/informers/externalversions/core/v1"
	apiv1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type CLEGController struct {
	ClusterName          string
	Cluster              *astrov1.Cluster
	SubExternalIp        string
	HeartbeatFrequency   time.Duration
	LeaseDurationSeconds time.Duration
	ForceSyncFrequency   time.Duration
	NamespacePrefix      string

	PlanetApiServerProtocol string
	PlanetApiServerPort     int32

	clusterInformer informersastrov1.ClusterInformer

	queue workqueue.RateLimitingInterface
}

func NewCLEGController(
	clusterName string,
	cluster *astrov1.Cluster,
	subExternalIp string,
	heartbeatFrequency time.Duration,
	leaseDurationSeconds time.Duration,
	forceSyncFrequency time.Duration,
	namespacePrefix string,
	planetApiServerProtocol string,
	planetApiServerPort int32,
	informer informersastrov1.ClusterInformer) *CLEGController {
	c := CLEGController{
		ClusterName:             clusterName,
		Cluster:                 cluster,
		SubExternalIp:           subExternalIp,
		HeartbeatFrequency:      heartbeatFrequency,
		LeaseDurationSeconds:    leaseDurationSeconds,
		ForceSyncFrequency:      forceSyncFrequency,
		NamespacePrefix:         namespacePrefix,
		PlanetApiServerPort:     planetApiServerPort,
		PlanetApiServerProtocol: planetApiServerProtocol,
		clusterInformer:         informer,
	}

	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CLEG-controller-pod-queue")
	cache.SubSharedInformerFactory.Core().V1().Pods().Informer().AddEventHandler(
		toolscache.ResourceEventHandlerFuncs{
			AddFunc:    c.addPodEvent,
			UpdateFunc: c.updatePodEvent,
			DeleteFunc: c.deletePodEvent,
		})

	cache.SubSharedInformerFactory.Core().V1().Nodes().Informer().AddEventHandler(
		toolscache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				node, ok := obj.(*apiv1.Node)
				if !ok {
					return false
				}
				if utils.GetSubClusterNodeLevel(node) > 0 {
					return false
				}
				return true
			},
			Handler: toolscache.ResourceEventHandlerFuncs{
				AddFunc:    c.addNodeEvent,
				UpdateFunc: c.updateNodeEvent,
				DeleteFunc: c.deleteNodeEvent,
			},
		})
	if c.clusterInformer != nil {
		c.clusterInformer.Informer().AddEventHandler(toolscache.ResourceEventHandlerFuncs{
			AddFunc:    c.addClusterEvent,
			UpdateFunc: c.updateClusterEvent,
			DeleteFunc: c.deleteNodeEvent,
		})
	}
	return &c
}

type event struct {
	eventType string
	name      string
}

func (c *CLEGController) addPodEvent(obj interface{}) {
	pod := obj.(*apiv1.Pod)
	key, err := toolscache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		klog.Error("addPodEvent error: %s", err)
		return
	}
	klog.V(4).Infof("add pod: %s", key)
	c.queue.Add(&event{eventType: "Pod", name: key})
}

func (c *CLEGController) updatePodEvent(old, new interface{}) {
	oldPod := old.(*apiv1.Pod)
	newPod := new.(*apiv1.Pod)
	if oldPod.Spec.NodeName == newPod.Spec.NodeName && oldPod.Status.Phase == newPod.Status.Phase {
		return
	}
	key, err := toolscache.MetaNamespaceKeyFunc(newPod)
	if err != nil {
		klog.Errorf("updatePodEvent: %s", err)
		return
	}
	klog.V(4).Infof("update pod: %s", key)
}

func (c *CLEGController) deletePodEvent(obj interface{}) {
	pod := obj.(*apiv1.Pod)
	key, err := toolscache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		klog.Errorf("deletePodEvent error: %s", err)
		return
	}
	klog.V(4).Infof("delete pod: %s", key)
	c.queue.Add(&event{eventType: "Pod", name: key})
}

func (c *CLEGController) addNodeEvent(obj interface{}) {
	node := obj.(*apiv1.Node)
	klog.V(4).Infof("add node: %s", node.Name)
	c.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (c *CLEGController) updateNodeEvent(old, new interface{}) {
	node := new.(*apiv1.Node)
	klog.V(4).Infof("update node: %s", node.Name)
	c.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (c *CLEGController) deleteNodeEvent(obj interface{}) {
	node := obj.(*apiv1.Node)
	klog.V(4).Infof("delete node: %s", node.Name)
	c.queue.Add(&event{eventType: "Node", name: node.Name})
}

func (c *CLEGController) addClusterEvent(obj interface{}) {
	cluster := obj.(*astrov1.Cluster)
	klog.V(4).Infof("add cluster: %s", cluster.Name)
	c.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}

func (c *CLEGController) updateClusterEvent(old, new interface{}) {
	cluster := new.(*astrov1.Cluster)
	klog.V(4).Infof("update cluster: %s", cluster.Name)
	c.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}

func (c *CLEGController) deleteClusterEvent(obj interface{}) {
	cluster := obj.(*astrov1.Cluster)
	klog.V(4).Infof("delete cluster: %s", cluster.Name)
	c.queue.Add(&event{eventType: "Cluster", name: cluster.Name})
}
