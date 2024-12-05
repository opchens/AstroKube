package sub_cluster

import (
	"context"
	"k8s.io/klog/v2"
	"time"
)

type SubManager struct {
	SubExternalIp      string
	ClusterName        string
	NamespacePrefix    string
	HeartBeatFrequency time.Duration

	SubApiServerProtocl string
	SubApiServerPort    int32
}

func (m SubManager) Run(ctx context.Context) {
	klog.Info("starting sub-cluster manager ... ")

	//var subClusterInformer informersastrocorev1.ClusterInformer
	//if cache.SubClientCache.ClusterLister != nil {
	//	subClusterInformer = cache.SubAstroSharedInformerFactory.Astro().V1().Clusters()
	//}
	//cluster, err := cache.CoreClientCache.ClusterLister.Clusters("").Get(m.ClusterName)
	//if err != nil {
	//	klog.Fatal("get cluster error: %s", err)
	//}
	//klog.Info("starting core-cluster controllers ... ")
	//go controller.NewClusterController(m.ClusterName, cluster, m.SubExternalIp, m.HeartBeatFrequency,
	//	m.NamespacePrefix, m.SubApiServerProtocl, m.SubApiServerPort, subClusterInformer).Run(ctx)
}
