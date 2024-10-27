package options

import "time"

type ServerRunOptions struct {
	CoreClusterKubeConfigPath string
	SubClusterKubeConfigPath  string
	ClusterName               string
	KubeClusterDomain         string
	FedNamespacePrefix        string
	SubClusterExternalIp      string
	Mode                      string

	LeaderElect                  bool
	LeaderElectLeaseDuration     time.Duration
	LeaderElectRenewDeadline     time.Duration
	LeaderElectRetryPeriod       time.Duration
	LeaderElectResourceLock      string
	LeaderElectResourceName      string
	LeaderElectResourceNamespace string

	KubeApiQPSForSub    float32
	KubeApiBurstForSub  int
	KubeApiQPSForCore   float32
	KubeApiBurstForCore int
}

func NewServerRunOptions() ServerRunOptions {
	return ServerRunOptions{}
}
