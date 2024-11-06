package astrolet

import (
	"AstroKube/pkg/astrolet/cache"
	"AstroKube/pkg/astrolet/core_cluster"
	"AstroKube/pkg/astrolet/sub_cluster"
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
	InformerResyncPeriod time.Duration
	ClusterName          string
	SubExternalIP        string
	MetricsAddr          string
	ListenPort           int32
	Mode                 string
}

func (as *AstroLet) Run(ctx context.Context) {
	cache.InitClientCache(ctx, as.SubClient, as.SubClient, as.InformerResyncPeriod, as.ClusterName, as.CoreConfig, as.SubConfig)
}

func (as *AstroLet) newSubManager(ctx context.Context) *core_cluster.CoreManager {
	return &core_cluster.CoreManager{}
}

func (as *AstroLet) newSubClusterManager(ctx context.Context) *sub_cluster.SubManager {
	return &sub_cluster.SubManager{}
}

func (as *AstroLet) startHttpServer(ctx context.Context) {}
