package controller

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/common"
	"AstroKube/pkg/astrolet/utils"
	informersastrov1 "AstroKube/pkg/client/informers/externalversions/core/v1"
	"context"
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	toolscache "k8s.io/client-go/tools/cache"
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

	SubApiServerProtocol string
	SubApiServerPort     int32

	clusterDaemonEndpoint astrov1.ClusterDaemonEndpoints
	asConfiguration       *astrov1.AstroletConfiguration
	apiServerAddress      []astrov1.Address
	clusterInfo           astrov1.Info
	secretRef             astrov1.ClusterSecretRef

	clusterInformer informersastrov1.ClusterInformer

	clusters map[string]*astrov1.Cluster

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

}

func (c *CLEGController) init(ctx context.Context) error {
	c.asConfiguration = c.initializeAsConfiguration(ctx)
	c.apiServerAddress = c.initializeAsAddress(ctx)
	c.clusterDaemonEndpoint = c.initializeDaemonEndpoint(ctx)
	c.clusterInfo = *c.initializeClusterInfo(ctx)
	c.secretRef = c.initializeSecretRef()

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

func (c *CLEGController) initializeClusterInfo(ctx context.Context) *astrov1.Info {
	version, err := cache.SubClientCache.Client.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("failed to get cluster version info")
		return nil
	}
	return &astrov1.Info{
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
