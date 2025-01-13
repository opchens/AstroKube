package controller

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/common"
	"AstroKube/pkg/astrolet/utils"
	informersastrov1 "AstroKube/pkg/client/informers/externalversions/core/v1"
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"net"
	"os"
	"strconv"
	"strings"
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
	TopNodeN             int

	SubApiServerProtocol string
	SubApiServerPort     int32

	clusterDaemonEndpoint astrov1.ClusterDaemonEndpoints
	asConfiguration       *astrov1.AstroletConfiguration
	apiServerAddress      []astrov1.Address
	clusterInfo           astrov1.ClusterInfo
	secretRef             astrov1.ClusterSecretRef

	clusterInformer informersastrov1.ClusterInformer

	nodes      map[string]*utils.NodeInfo
	clusters   map[string]*astrov1.Cluster
	pods       map[string]*apiv1.Pod
	scoreNodes utils.ScoreNodeList

	allocatable *utils.Resource
	usage       *utils.Resource

	subClusterAllocatable *utils.Resource
	subClusterUsage       *utils.Resource

	queue workqueue.RateLimitingInterface

	ignoredCustomResources []string
}

func NewCLEGController(
	clusterName string,
	cluster *astrov1.Cluster,
	subExternalIp string,
	heartbeatFrequency time.Duration,
	leaseDurationSeconds time.Duration,
	forceSyncFrequency time.Duration,
	namespacePrefix string,
	subApiServerProtocol string,
	subApiServerPort int32,
	informer informersastrov1.ClusterInformer) *CLEGController {
	c := CLEGController{
		ClusterName:          clusterName,
		Cluster:              cluster,
		SubExternalIp:        subExternalIp,
		HeartbeatFrequency:   heartbeatFrequency,
		LeaseDurationSeconds: leaseDurationSeconds,
		ForceSyncFrequency:   forceSyncFrequency,
		NamespacePrefix:      namespacePrefix,
		SubApiServerPort:     subApiServerPort,
		SubApiServerProtocol: subApiServerProtocol,
		clusterInformer:      informer,
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

func (c *CLEGController) Run(ctx context.Context) {
	utils.ContextStart(ctx)
	defer utils.ContextStop(ctx)
	klog.Infof("starting CELG-controller")
	if err := c.init(ctx); err != nil {
		klog.Fatalf("cannot init, error: %v", err)
	}
	if err := c.ensureCluster(ctx); err != nil {
		klog.Errorf("cannot ensure if cluster is exist, error: %v", err)
	}
	if err := c.ensureNamespace(ctx); err != nil {
		klog.Errorf("cannot ensure if namespace is exist, error: %v", err)
	}

	go wait.UntilWithContext(ctx, c.updateStatus, c.ForceSyncFrequency)
	go wait.UntilWithContext(ctx, c.syncClusterLabelToNode, c.ForceSyncFrequency)

	c.sync()
}

func (c *CLEGController) sync() {
	for {
		key, shutdown := c.queue.Get()
		if shutdown {
			return
		}
		e := key.(*event)
		klog.V(4).Infof("start reconcile %s %s", e.eventType, e.name)
		var err error
		switch e.eventType {
		case "Pod":
			err = c.syncPod(e)
		case "Node":
			err = c.syncNode(e)
		case "Cluster":
			err = c.syncCluster(e)
		default:
			klog.Warningf("got unknown eventType: %s", e.eventType)
		}
		if err != nil {
			klog.Errorf("sync error %s/%s: $%s", e.eventType, e.name, err)
			c.queue.AddRateLimited(e)
			return
		} else {
			c.queue.Forget(e)
		}
		c.queue.Done(e)
		klog.V(4).Infof("finish reconcile %s %s", e.eventType, e.name)
	}
}

func (c *CLEGController) syncPod(e *event) error {
	klog.V(4).Infof("start sync pod %s", e.name)
	namespace, name, err := toolscache.SplitMetaNamespaceKey(e.name)
	if err != nil {
		klog.Errorf("syncPod %s error: %s", e, err)
		return nil
	}
	pod, err := cache.SubClientCache.PodLister.Pods(namespace).Get(name)
	if errors.IsNotFound(err) {
		cachePod, ok := c.pods[e.name]
		if ok {
			c.deletePodFromCache(cachePod, e)
		}
		return nil
	} else if err != nil {
		return err
	}
	if _, ok := pod.Annotations["location"]; !ok {
		return nil
	}
	// update pod cache if phase is not podsuccessed or podfailed
	if cachePod, ok := c.pods[e.name]; ok {
		if pod.Status.Phase == apiv1.PodSucceeded || pod.Status.Phase == apiv1.PodFailed {
			klog.V(4).Infof("pod.status.phase Successed or Failed, remove pod %s from cache", e.name)
			c.deletePodFromCache(cachePod, e)
		} else {
			klog.V(4).Infof("update  pod %s in cache", e.name)
			c.pods[e.name] = pod.DeepCopy()
		}
	} else if pod.Spec.NodeName != "" && pod.Status.Phase != apiv1.PodSucceeded && pod.Status.Phase != apiv1.PodFailed {
		// pod not in cache and has scheduled
		klog.V(4).Infof("pod %s been scheduled to node %s", e.name, pod.Spec.NodeName)
		nodeInfo, ok := c.nodes[pod.Spec.NodeName]
		if ok {
			// add usage cache
			c.pods[e.name] = pod.DeepCopy()
			res := nodeInfo.AddPod(pod)
			c.usage.AddResource(res)
			patch := []string{"usage"}
			nodeInfo.Score()
			originIndex := nodeInfo.Index
			c.scoreNodes.Down(nodeInfo.Index)
			if originIndex < c.TopNodeN {
				patch = append(patch, "nodes")
			}
			err := c.patchStatus(patch)
			if err != nil {
				klog.Errorf("syncPod: patchStatus error: %v", err)
			}
		} else {
			klog.Warningf("syncPod get pod %s update event, pod.Sepc.NodeName is %s, but not found node in cc.nodes", e.name, pod.Spec.NodeName)
		}
	} else {
		klog.V(4).Infof("pod %s has not been scheduled", e.name)
	}
	return nil
}

func (c *CLEGController) syncNode(e *event) error {
	klog.V(4).Infof("start sync node %s", e.name)
	node, err := cache.SubClientCache.NodeLister.Get(e.name)
	if errors.IsNotFound(err) || !utils.IsNodeReady(node) {
		// node deleted or notready
		nodeInfo, ok := c.nodes[e.name]
		if ok {
			// delete cache
			delete(c.nodes, e.name)
			c.usage.SubResource(nodeInfo.Usage)
			c.allocatable.SubResource(nodeInfo.Allocatable)
			for name, pod := range c.pods {
				if pod.Spec.NodeName == e.name {
					delete(c.pods, name)
				}
			}
			patch := []string{"usage", "allocatable"}
			c.scoreNodes.Remove(nodeInfo.Index)
			if nodeInfo.Index < c.TopNodeN {
				patch = append(patch, "nodes")
			}
			err := c.patchStatus(patch)
			if err != nil {
				klog.Errorf("syncNode: patchStatus error: %v", err)
			}
		} else {
			klog.Warningf("syncNode get node %s delete event, but not found node in c.nodes", e.name)
		}
		return nil
	} else if err != nil {
		return err
	}
	nodeInfo, ok := c.nodes[e.name]
	if ok {
		klog.V(4).Infof("node %s has been update", e.name)
		cacheNodeInfo := &utils.NodeInfo{
			Node:        nodeInfo.Node,
			Allocatable: nodeInfo.Allocatable.DeepCopy(),
			Usage:       nodeInfo.Usage.DeepCopy(),
			Pods:        nodeInfo.Pods,
		}
		nodeInfo.Node = node
		allocatable := utils.NewResource(node.Status.Allocatable)
		var diff *utils.Resource
		if !allocatable.Equal(cacheNodeInfo.Allocatable) {
			diff = allocatable.Diff(cacheNodeInfo.Allocatable)
			nodeInfo.Allocatable = allocatable.DeepCopy()
			c.allocatable.AddResource(diff)
			patch := []string{"allocatable"}
			nodeInfo.Score()
			listChanged := c.scoreNodes.Up(nodeInfo.Index)
			if listChanged {
				if nodeInfo.Index < c.TopNodeN {
					patch = append(patch, "nodes")
				}
			} else {
				originIndex := nodeInfo.Index
				c.scoreNodes.Down(nodeInfo.Index)
				if originIndex < c.TopNodeN {
					patch = append(patch, "nodes")
				}
			}
			err := c.patchStatus(patch)
			if err != nil {
				klog.Errorf("syncNode: patchStatus error: %v", err)
			}
		}
	} else {
		klog.V(4).Infof("node %s has been added", e.name)
		nodeInfo := utils.NewNodeInfo(node)
		nodeInfo.Index = c.scoreNodes.Len()
		c.nodes[node.Name] = nodeInfo
		c.allocatable.AddResource(nodeInfo.Allocatable)
		pods, _ := cache.SubClientCache.PodLister.List(labels.Everything())
		for _, pod := range pods {
			if pod.Spec.NodeName == e.name {
				key, err := toolscache.MetaNamespaceKeyFunc(pod)
				if err != nil {
					klog.Errorf("addPodEvent error: %s", err)
					continue
				}
				klog.V(4).Infof("add pod: %s", key)
				c.queue.Add(&event{"Pod", key})
			}
		}
		patch := []string{"allocatable"}
		nodeInfo.Score()
		c.scoreNodes = append(c.scoreNodes, nodeInfo)
		c.scoreNodes.Up(nodeInfo.Index)
		if nodeInfo.Index < c.TopNodeN {
			patch = append(patch, "nodes")
		}
		err = c.patchStatus(patch)
		if err != nil {
			klog.Errorf("syncNode: patchStatus error: %v", err)
		}
	}
	return nil
}

func (c *CLEGController) syncCluster(e *event) error {
	cluster, err := cache.SubClientCache.ClusterLister.Clusters("default").Get(e.name)
	if errors.IsNotFound(err) {
		// deleted sub cluster in core cluster
		nameInCore := utils.ConvertClusterName(c.ClusterName, e.name)
		err = cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").Delete(context.TODO(), nameInCore, metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			klog.Infof("syncCluster: cluster %s has been deleted", nameInCore)
		} else if err != nil {
			return err
		}
		cacheCluster, ok := c.clusters[e.name]
		if ok {
			delete(c.clusters, e.name)
			c.subClusterUsage.SubResource(utils.NewResource(cacheCluster.Status.SubClusterUsage))
			c.subClusterAllocatable.SubResource(utils.NewResource(cacheCluster.Status.Allocatable))
			err := c.patchStatus([]string{"subClusterUsage", "subClusterAllocatable", "subClusterNodes"})
			if err != nil {
				klog.Errorf("syncCluster: patchStatus error: %v", err)
			}
		} else {
			klog.Warningf("syncCluster get cluster %s deleted event, but not found in c.clusters", e.name)
		}
		return nil
	} else if err != nil {
		return err
	}
	// update sub cluster in core cluster
	clusterInCore := utils.ConvertCluster(c.ClusterName, cluster)
	err = createOrUpdateCluster(clusterInCore, cluster)
	if err != nil {
		return err
	}
	if cluster.Labels[astrov1.ClusterLevelLabel] != "1" {
		return nil
	}
	cacheCluster, ok := c.clusters[e.name]
	if ok {
		if cluster.Status.Phase == astrov1.ONLINE {
			c.clusters[e.name] = cluster.DeepCopy()
			changed := []string{"subClusterNodes"}
			subClusterUsage := utils.NewResource(cluster.Status.SubClusterUsage)
			cacheSubClusterUsage := utils.NewResource(cacheCluster.Status.SubClusterUsage)
			if !subClusterUsage.Equal(cacheSubClusterUsage) {
				subClusterUsage.SubResource(cacheSubClusterUsage)
				c.subClusterUsage.AddResource(subClusterUsage)
				changed = append(changed, "subClusterUsage")
			}
			subClusterAllocatable := utils.NewResource(cluster.Status.SubClusterAllocatable)
			cacheSubClusterAllocatable := utils.NewResource(cacheCluster.Status.SubClusterAllocatable)
			if !subClusterAllocatable.Equal(cacheSubClusterAllocatable) {
				subClusterAllocatable.SubResource(cacheSubClusterAllocatable)
				c.subClusterAllocatable.AddResource(subClusterAllocatable)
				changed = append(changed, "subClusterAllocatable")
			}
			err = c.patchStatus(changed)
		} else {
			klog.V(4).Infof("cluster %s offline", e.name)
			delete(c.clusters, e.name)
			c.subClusterUsage.SubResource(utils.NewResource(cacheCluster.Status.SubClusterUsage))
			c.subClusterAllocatable.SubResource(utils.NewResource(cacheCluster.Status.SubClusterAllocatable))
			err = c.patchStatus([]string{"subClusterUsage", "subClusterAllocatable", "subClusterNodes"})
		}
	} else if cluster.Status.Phase == astrov1.ONLINE {
		klog.V(4).Infof("cluster %s added", e.name)
		c.clusters[e.name] = cluster.DeepCopy()
		c.subClusterUsage.AddResource(utils.NewResource(cluster.Status.SubClusterUsage))
		c.subClusterAllocatable.AddResource(utils.NewResource(cluster.Status.SubClusterAllocatable))
		err = c.patchStatus([]string{"subClusterUsage", "subClusterAllocatable", "subClusterNodes"})
	}
	if err != nil {
		klog.Errorf("syncCluster: patchStatus error: %v", err)
	}
	return nil
}

func (c *CLEGController) deletePodFromCache(cachePod *apiv1.Pod, e *event) {
	delete(c.pods, e.name)
	nodeInfo, ok := c.nodes[cachePod.Spec.NodeName]
	if ok {
		res := nodeInfo.RemovePod(cachePod)
		c.usage.SubResource(res)
		patch := []string{"usage"}
		nodeInfo.Score()
		c.scoreNodes.Up(nodeInfo.Index)
		if nodeInfo.Index > c.TopNodeN {
			patch = append(patch, "nodes")
		}
		err := c.patchStatus(patch)
		if err != nil {
			klog.Errorf("syncPod: patchStatus error: %v", err)
		}
	} else {
		klog.Warningf("syncPod get pod %s delete event, pod.Sepc.NodeName is %s, but not found node in cc.nodes", e.name, cachePod.Spec.NodeName)

	}
}

func (c *CLEGController) patchStatus(changes []string) error {
	if len(changes) == 0 {
		return nil
	}
	payloads := []utils.PatchStatus{}
	if c.clusterInformer == nil {
		addChanges := []string{}
		for _, change := range changes {
			switch change {
			case "usage":
				c.subClusterUsage = c.usage.DeepCopy()
				addChanges = append(addChanges, "subClusterUsage")
			case "allocatable":
				c.subClusterAllocatable = c.allocatable.DeepCopy()
				addChanges = append(addChanges, "subClusterAllocatable")
			case "nodes":
				payload := utils.PatchStatus{
					Op:    "replace",
					Path:  "/status/subClusterNode",
					Value: c.scoreNodes.Top(c.TopNodeN),
				}
				payloads = append(payloads, payload)
			}
		}
		changes = append(changes, addChanges...)
	}
	for _, change := range changes {
		switch change {
		case "usage":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/usage",
				Value: c.usage.ToResourceList(),
			}
			payloads = append(payloads, payload)
		case "allocatable":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/allocatable",
				Value: c.allocatable.ToResourceList(),
			}
			payloads = append(payloads, payload)
		case "nodes":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/nodes",
				Value: c.scoreNodes.Top(c.TopNodeN),
			}
			payloads = append(payloads, payload)
		case "subClusterUsage":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/subClusterUsage",
				Value: c.subClusterUsage.ToResourceList(),
			}
			payloads = append(payloads, payload)
		case "subClusterAllocatable":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/subClusterAllocatable",
				Value: c.subClusterAllocatable.ToResourceList(),
			}
			payloads = append(payloads, payload)
		case "subClusterNodes":
			payload := utils.PatchStatus{
				Op:    "replace",
				Path:  "/status/subClusterNodes",
				Value: c.getSubClusterNodes(),
			}
			payloads = append(payloads, payload)
		}
	}
	payloadBytes, err := json.Marshal(payloads)
	if err != nil {
		return err
	}
	starClient := cache.CoreClientCache.AstroClient.AstroV1().Clusters("default")
	klog.V(4).Infof("patch cluster status: %s", payloads)
	_, err = starClient.Patch(context.TODO(), c.ClusterName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{}, "status")
	return err
}

func (c *CLEGController) getSubClusterNodes() []astrov1.NodeLeftResource {
	ret := make([]astrov1.NodeLeftResource, 0, c.TopNodeN)
	clusterNodes := make(map[string]*utils.Resource)
	clusterIndex := make(map[string]int)
	for _, cluster := range c.clusters {
		if len(cluster.Status.SubClusterNodes) == 0 {
			clusterIndex[cluster.Name] = -1
		} else {
			clusterNodes[cluster.Name] = utils.NewResource(cluster.Status.SubClusterNodes[0].Left)
			clusterIndex[cluster.Name] = 0
		}
	}
	for i := 0; i < c.TopNodeN; i++ {
		clusterName := ""
		nodeName := ""
		max := utils.NewResource(nil)
		for _, cluster := range c.clusters {
			if clusterIndex[cluster.Name] == -1 {
				continue
			}
			if max.Less(clusterNodes[cluster.Name]) {
				clusterName = cluster.Name
				nodeName = cluster.Status.SubClusterNodes[clusterIndex[clusterName]].Name
				max = clusterNodes[cluster.Name]
			}
		}
		if clusterName == "" {
			break
		}
		ret = append(ret, astrov1.NodeLeftResource{
			Name: clusterName + "." + nodeName,
			Left: max.ToResourceList(),
		})
		clusterIndex[clusterName]++
		cluster := c.clusters[clusterName]
		index := clusterIndex[clusterName]
		if index >= len(cluster.Status.SubClusterNodes) {
			clusterIndex[clusterName] = -1
		} else {
			clusterNodes[clusterName] = utils.NewResource(cluster.Status.SubClusterNodes[index].Left)
		}
	}
	return ret
}

func (c *CLEGController) init(ctx context.Context) error {
	c.asConfiguration = c.initializeAsConfiguration(ctx)
	c.apiServerAddress = c.initializeAsAddress(ctx)
	c.clusterDaemonEndpoint = c.initializeDaemonEndpoint(ctx)
	c.clusterInfo = *c.initializeClusterInfo(ctx)
	c.secretRef = c.initializeSecretRef()

	nodes, err := cache.SubClientCache.NodeLister.List(labels.Everything())
	if err != nil {
		return err
	}
	c.nodes = make(map[string]*utils.NodeInfo)
	for _, node := range nodes {
		if utils.GetSubClusterNodeLevel(node) > 0 {
			continue
		}
		if utils.IsNodeReady(node) {
			c.nodes[node.Name] = utils.NewNodeInfo(node)
		}
	}

	pods, err := cache.SubClientCache.PodLister.List(labels.Everything())
	if err != nil {
		return err
	}
	c.pods = make(map[string]*apiv1.Pod)
	for _, pod := range pods {
		key, err := toolscache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		if pod.Spec.NodeName != "" && pod.Status.Phase != apiv1.PodSucceeded && pod.Status.Phase != apiv1.PodFailed {
			if nodeInfo, ok := c.nodes[pod.Spec.NodeName]; ok {
				nodeInfo.AddPod(pod)
				c.pods[key] = pod.DeepCopy()
			}
		}
	}

	index := 0
	for _, nodeInfo := range c.nodes {
		nodeInfo.Score()
		nodeInfo.Index = index
		index++
		c.usage.AddResource(nodeInfo.Usage)
		c.allocatable.AddResource(nodeInfo.Allocatable)
	}

	if c.clusterInformer != nil {
		clusters, err := cache.SubClientCache.ClusterLister.List(labels.Everything())
		if err != nil {
			return err
		}
		c.clusters = make(map[string]*astrov1.Cluster)
		for _, cl := range clusters {
			cluster := utils.ConvertCluster(c.ClusterName, cl)
			err = createOrUpdateCluster(cluster, cl)
			if err != nil {
				return err
			}
			if cl.Status.Phase != astrov1.ONLINE || cl.Labels[astrov1.ClusterLabel] != "1" {
				continue
			}
			c.subClusterUsage.AddResource(utils.NewResource(cl.Status.SubClusterUsage))
			c.subClusterAllocatable.AddResource(utils.NewResource(cl.Status.SubClusterAllocatable))
			c.clusters[cl.Name] = cl.DeepCopy()
		}
	}
	return nil
}

func (c *CLEGController) initializeSecretRef() astrov1.ClusterSecretRef {
	return astrov1.ClusterSecretRef{
		Name:      "cluster-credential-" + c.ClusterName,
		Namespace: common.AstroSystem,
	}
}

func (c *CLEGController) initializeAsConfiguration(ctx context.Context) *astrov1.AstroletConfiguration {
	return &astrov1.AstroletConfiguration{
		HeartbeatFrequency:   int64(c.HeartbeatFrequency.Seconds()),
		LeaseDurationSeconds: int64(c.LeaseDurationSeconds.Seconds()),
		ForceSyncFrequency:   int64(c.ForceSyncFrequency.Seconds()),
		NamespacePrefix:      c.NamespacePrefix,
	}
}

func (c *CLEGController) initializeAsAddress(ctx context.Context) []astrov1.Address {
	Svc, err := cache.SubClientCache.ServiceLister.Services("default").Get("kubernetes")
	if err != nil {
		klog.Warningln("get service/kubernetes failed")
	}
	return []astrov1.Address{
		{
			AddressIP: c.SubExternalIp,
			Type:      astrov1.ExternalIP,
		},
		{
			AddressIP: Svc.Spec.ClusterIP,
			Type:      astrov1.InternalIP,
		},
	}
}

func (c *CLEGController) initializeDaemonEndpoint(ctx context.Context) astrov1.ClusterDaemonEndpoints {
	protocol := os.Getenv("AS_PROTOCOL")
	if protocol != "https" && protocol != "http" {
		klog.Warningf("env AS_PROTOCL should be http ( default ) or https, got: %s", protocol)
		protocol = "http"
	}
	addr := os.Getenv("AS_ADDRESS")
	address := net.ParseIP(addr)
	if address == nil {
		klog.Warningf("env AS_ADDRESS should be an ip address, got: %s", address.String())
		var err error
		addr, err = GetLocalIp()
		if err != nil {
			klog.Warningf("get local ip failed: %v", err)
		}
	}
	portStr := os.Getenv("AS_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		klog.Errorf("env AS_PORT should be an integer between 1 to 65535, got: %s", portStr)
	}
	return astrov1.ClusterDaemonEndpoints{
		AstroletEndpoint: astrov1.DaemonEndpoint{
			Port:     int32(port),
			Protocol: protocol,
			Address:  addr,
		},
		ApiServerEndpoint: astrov1.DaemonEndpoint{
			Port:     c.SubApiServerPort,
			Protocol: c.SubApiServerProtocol,
			Address:  c.SubExternalIp,
		},
	}
}

func (c *CLEGController) initializeClusterInfo(ctx context.Context) *astrov1.ClusterInfo {
	version, err := cache.SubClientCache.Client.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("failed to get cluster version info")
		return nil
	}
	return &astrov1.ClusterInfo{
		Major:        version.Major,
		Minor:        version.Minor,
		GitVersion:   version.GitVersion,
		GitCommit:    version.GitCommit,
		GitTreeState: version.GitTreeState,
		BuildDate:    version.BuildDate,
		GoVersion:    version.GoVersion,
		Compiler:     version.Compiler,
		Platform:     version.Platform,
	}
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func createOrUpdateCluster(cluster *astrov1.Cluster, clusterInSub *astrov1.Cluster) error {
	klog.V(4).Infof("sub cluster %s update ...", cluster.Name)
	clusterInCore, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(cluster.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// create cluster to core
			cluster.Labels[astrov1.ClusterFullNameLabel] = cluster.Name
			//  resourceVersion should not be set on objects to be created
			cluster.UID = ""
			cluster.ResourceVersion = ""
			cluster.SelfLink = ""
			_, err := cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").Create(context.TODO(), cluster, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("create cluster in core error: %s", err)
				return err
			}
		} else {
			klog.Errorf("get core cluster error: %s", err)
			return err
		}
	} else {
		fullName := cluster.Name
		if fnInCore, ok := clusterInCore.Labels[astrov1.ClusterFullNameLabel]; ok && strings.HasSuffix(fnInCore, fullName) {
			fullName = fnInCore
		}
		if fnInSub, ok := clusterInSub.Labels[astrov1.ClusterFullNameLabel]; !ok || fnInSub != fullName {
			clusterCopy := clusterInSub.DeepCopy()
			clusterCopy.Labels[astrov1.ClusterFullNameLabel] = fullName
			_, err := cache.SubClientCache.AstroClient.AstroV1().Clusters("default").Update(context.TODO(), clusterCopy, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("update cluster error: %v", err)
				return err
			}
		}
		//update core cluster
		cluster.Labels[astrov1.ClusterFullNameLabel] = fullName
		clusterCopy := cluster.DeepCopy()
		clusterCopy.Status = *clusterInCore.Status.DeepCopy()
		if !apiequality.Semantic.DeepEqual(clusterInCore, clusterCopy) {
			cluster.UID = clusterInCore.UID
			cluster.ResourceVersion = clusterInCore.ResourceVersion
			_, err := cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").Update(context.TODO(), cluster, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("update core cluster error: %v", err)
				return err
			}
		}
	}
	cluster.Status.Phase = astrov1.ONLINE
	if !apiequality.Semantic.DeepEqual(clusterInCore.Status, cluster.Status) {
		_, err := cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("updateStatus core cluster error: %v", err)
			return err
		}
	}
	return nil
}

func (c *CLEGController) ensureCluster(ctx context.Context) error {
	cluster, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
	retryFlag := false
	if err == nil {
		retryErr := utils.RetryConflictOrServiceUnavailable(retry.DefaultBackoff, func() error {
			var errIn error
			if retryFlag == true {
				cluster, errIn = cache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
				if errIn != nil {
					return errIn
				}
			}
			clusterCopy := cluster.DeepCopy()
			clusterCopy.Labels = make(map[string]string)
			clusterCopy.Labels[astrov1.ClusterLevelLabel] = "1"
			clusterCopy.Labels[astrov1.ClusterFullNameLabel] = c.ClusterName
			if fullname, ok := clusterCopy.Labels[astrov1.ClusterFullNameLabel]; !ok || !strings.HasSuffix(fullname, c.ClusterName) {
				clusterCopy.Labels[astrov1.ClusterFullNameLabel] = c.ClusterName
			}
			if !apiequality.Semantic.DeepEqual(clusterCopy, cluster) {
				_, err = cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").Update(ctx, clusterCopy, metav1.UpdateOptions{})
				if err != nil {
					retryFlag = true
					return err
				}
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	} else if errors.IsNotFound(err) {
		newCluster := c.newCluster(ctx)
		_, err = cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").Create(ctx, newCluster, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			klog.Errorf("register cluster failed, error: ", err.Error())
			return err
		}
	} else {
		return err
	}

	ignoredCR, err := cache.SubClientCache.Client.CoreV1().ConfigMaps(common.AstroSystem).Get(context.TODO(), common.IgnoredCustomResourcesConfigMap, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			ignoredCR = &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: common.AstroSystem,
					Name:      common.IgnoredCustomResourcesConfigMap,
				},
				Data: map[string]string{
					common.IgnoredCustomResourcesConfigMap: strings.Join(common.IgnoredCustomResources, "\n"),
				},
			}
			ignoredCR, err = cache.SubClientCache.Client.CoreV1().ConfigMaps(common.AstroSystem).Create(context.TODO(), ignoredCR, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create configmap %s/%s in sub cluster: %v", common.AstroSystem, common.IgnoredCustomResourcesConfigMap, err)
			}
		} else {
			return fmt.Errorf("failed to get configmap %s/%s in sub cluster: %v", common.AstroSystem, common.IgnoredCustomResourcesConfigMap, err)
		}
	}
	for _, s := range strings.Split(ignoredCR.Data[common.IgnoredCustomResourcesData], "\n") {
		s = strings.TrimSpace(s)
		if len(s) != 0 {
			c.ignoredCustomResources = append(c.ignoredCustomResources, s)
		}
	}
	return utils.RetryConflictOrServiceUnavailable(retry.DefaultBackoff, func() error {
		cluster, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
		if err != nil {
			return err
		}
		clusterCopy := cluster.DeepCopy()
		clusterCopy.Status.ClusterInfo = c.clusterInfo
		clusterCopy.Status.Configuration = c.asConfiguration
		clusterCopy.Status.Addresses = c.apiServerAddress
		clusterCopy.Status.DaemonEndpoint = c.clusterDaemonEndpoint
		clusterCopy.Status.SecretRef = c.secretRef
		if c.clusterInformer == nil {
			clusterCopy.Status.SubClusterNodes = clusterCopy.Status.Nodes
		}
		if !apiequality.Semantic.DeepEqual(clusterCopy.Status, cluster.Status) {
			_, err = cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").UpdateStatus(ctx, clusterCopy, metav1.UpdateOptions{})
			return err
		}
		return nil
	})
}

func (c *CLEGController) ensureNamespace(ctx context.Context) error {
	_, err := cache.CoreClientCache.NamespaceLister.Get(common.AstroSystem)
	if err != nil {
		if errors.IsNotFound(err) {
			ns := &apiv1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: common.AstroSystem,
				},
			}

			_, err := cache.CoreClientCache.Client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create namespace/%s in core cluster: %v", common.AstroSystem, err)
			}
		} else {
			return fmt.Errorf("failed to get namespace/%s in core cluster: %v", common.AstroSystem, err)
		}
	}
	_, err = cache.SubClientCache.NamespaceLister.Get(common.AstroSystem)
	if err != nil {
		if errors.IsNotFound(err) {
			ns := &apiv1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: common.AstroSystem,
				},
			}
			_, err := cache.SubClientCache.Client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create namespace/%s in planet cluster: %v", common.AstroSystem, err)
			}
		} else {
			return fmt.Errorf("failed to get namespace/mizargalaxy-system in planet cluster: %v", err)
		}
	}
	return nil
}

func (c *CLEGController) updateStatus(ctx context.Context) {
	klog.V(4).Infof("start update status")
	defer klog.V(4).Infof("finish update status")
	cluster, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
	if err != nil {
		klog.Errorf("failed to get cluster %s:%s", c.ClusterName, err)
		return
	}
	clusterCopy := cluster.DeepCopy()
	clusterCopy.Status.Usage = c.usage.ToResourceList()
	clusterCopy.Status.Allocatable = c.allocatable.ToResourceList()
	clusterCopy.Status.Nodes = c.scoreNodes.Top(c.TopNodeN)
	klog.V(4).Infof("update cluster status: %s", clusterCopy.Status)
	_, err = cache.CoreClientCache.AstroClient.AstroV1().Clusters("default").UpdateStatus(ctx, clusterCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update cluster %s status: %s", c.ClusterName, err)
	}
}

func (c *CLEGController) syncClusterLabelToNode(ctx context.Context) {
	cluster, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(c.ClusterName)
	if err != nil {
		klog.Errorf("failed to get cluster %s", c.ClusterName)
	}
	klog.Infof("Start to syncClusterLbaelToNode for cluster %s", c.ClusterName)
	if cluster.Labels == nil || cluster.Labels[astrov1.LabelClusterName] == "" {
		return
	}
	if cluster.Labels[astrov1.LabelClusterName] != cluster.Name {
		klog.Warningf("cluster label %s value %s not equal cluster name %s", astrov1.LabelClusterName, cluster.Labels[astrov1.LabelClusterName], c.ClusterName)
		return
	}
	nodes, err := cache.SubClientCache.NodeLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to list the nodes in cluster %s: %v", cluster.Name, err)
		return
	}
	for _, node := range nodes {
		if utils.GetSubClusterNodeLevel(node) > 0 {
			continue
		}
		if err := c.syncNodeLabel(node, cluster.Labels); err != nil {
			klog.Errorf("failed to update the node %s in cluster %s labels: %v", node.Name, cluster.Name, err)
		}
	}
}

func (c *CLEGController) syncNodeLabel(node *apiv1.Node, label map[string]string) error {
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	nodeCopy := node.DeepCopy()
	nodeCopy.Labels[astrov1.LabelClusterName] = c.ClusterName
	for _, topoKey := range utils.ClusterTopoKeyList {
		val, ok := label[topoKey]
		if !ok {
			_, ok1 := nodeCopy.Labels[topoKey]
			if ok1 {
				delete(nodeCopy.Labels, topoKey)
			}
		} else {
			if val != nodeCopy.Labels[topoKey] {
				nodeCopy.Labels[topoKey] = val
			}
		}
	}
	if !apiequality.Semantic.DeepEqual(node.Labels, nodeCopy.Labels) {
		_, err := cache.SubClientCache.Client.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func (c *CLEGController) newCluster(ctx context.Context) *astrov1.Cluster {
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
