package astrolet

import (
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/core_cluster"
	"AstroKube/pkg/astrolet/sub_cluster"
	"AstroKube/pkg/client/clientset/versioned"
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

type AstroLet struct {
	CoreClient           kubernetes.Interface
	CoreConfig           *rest.Config
	SubClient            kubernetes.Interface
	SubConfig            *rest.Config
	SubAsClient          versioned.Interface
	CoreAsClient         versioned.Interface
	InformerResyncPeriod time.Duration
	HeartbeatFrequency   time.Duration
	ForceSyncFrequency   time.Duration
	ClusterName          string
	SubExternalIP        string
	MetricsAddr          string
	ListenPort           int32
	NamespacePrefix      string
	Mode                 string
}

func (as *AstroLet) Run(ctx context.Context) {
	cache.InitClientCache(ctx, as.SubClient, as.SubClient, as.SubAsClient, as.CoreAsClient, as.InformerResyncPeriod,
		as.ClusterName, as.CoreConfig, as.SubConfig)
	go as.newCoreManager().Run(ctx)
	go as.newSubClusterManager().Run(ctx)
	go as.startHttpServer(ctx)

}

func (as *AstroLet) newCoreManager() *core_cluster.CoreManager {
	return &core_cluster.CoreManager{}
}

func (as *AstroLet) newSubClusterManager() *sub_cluster.SubManager {
	return &sub_cluster.SubManager{
		ClusterName:        as.ClusterName,
		NamespacePrefix:    as.NamespacePrefix,
		ForceSyncFrequency: as.ForceSyncFrequency,
	}
}

func (as *AstroLet) startHttpServer(ctx context.Context) {
	ctx.Done()
}
