package astrolet

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type AstroLet struct {
	CoreClient      kubernetes.Interface
	CoreConfig      *rest.Config
	SubClient       kubernetes.Interface
	SubClientConfig *rest.Config
	SubExternalIP   string
	MetricsAddr     string
	ListenPort      int32
	Mode            string
}

func (as *AstroLet) Run(ctx context.Context) {
	//  Run astrolet server
}
