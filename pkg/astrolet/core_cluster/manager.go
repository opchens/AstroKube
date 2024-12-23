package core_cluster

import (
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/core_cluster/controller"
	informersastrov1 "AstroKube/pkg/client/informers/externalversions/core/v1"
	"context"
	"k8s.io/klog/v2"
	"time"
)

type CoreManager struct {
	ExternalIp           string
	ClusterName          string
	HeartbeatFrequency   time.Duration
	LeaseDurationSeconds time.Duration
	ForceSyncFrequency   time.Duration
	NamespacePrefix      string

	PlanetApiServerProtocol string
	PlanetApiServerPort     int32
}

func (m CoreManager) Run(ctx context.Context) {
	klog.Info("starting core-cluster manager")
	var clusterInformer informersastrov1.ClusterInformer
	if cache.SubClientCache.AstroClient != nil {
		clusterInformer = cache.SubAstroSharedInformerFactory.Astro().V1().Clusters()
	}
	cluster, err := cache.CoreClientCache.ClusterLister.Clusters("default").Get(m.ClusterName)
	if err != nil {
		klog.Errorf("get sub cluster error: %v", err)
	}
	CLEGController := controller.NewCLEGController(
		m.ClusterName,
		cluster,
		m.ExternalIp,
		m.HeartbeatFrequency,
		m.LeaseDurationSeconds,
		m.ForceSyncFrequency,
		m.NamespacePrefix,
		m.PlanetApiServerProtocol,
		m.PlanetApiServerPort,
		clusterInformer,
	)
	go CLEGController.Run(ctx)
}
