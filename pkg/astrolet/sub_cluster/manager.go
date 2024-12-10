package sub_cluster

import (
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/sub_cluster/controller"
	"context"
	"k8s.io/klog/v2"
	"time"
)

type SubManager struct {
	ClusterName     string
	NamespacePrefix string

	ForceSyncFrequency time.Duration
}

func (m SubManager) Run(ctx context.Context) {
	klog.Info("starting sub-cluster manager ... ")
	go controller.NewComponentsController(cache.SubClientCache.Client, m.ClusterName, m.ForceSyncFrequency).Run(ctx)
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
