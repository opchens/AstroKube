package options

import "time"

const (
	DefaultClusterName        = "astro-sub"
	DefaultFedNamespacePrefix = "ast"
	DefaultMode               = "normal"
	DefaultKubeClusterDomain  = "cluster.local"
)

type ServerRunOptions struct {
	CoreClusterKubeConfigPath string
	SubClusterKubeConfigPath  string
	ClusterName               string
	KubeClusterDomain         string
	FedNamespacePrefix        string
	SubClusterExternalIp      string
	Mode                      string

	InformerResyncPeriod time.Duration
	LeaseDurationSeconds time.Duration
	HeartbeatFrequency   time.Duration
	ForceSyncFrequency   time.Duration

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

func SetDefaultOpts(opts *ServerRunOptions) error {
	if opts.ClusterName == "" {
		opts.ClusterName = DefaultClusterName
	}
	if opts.KubeClusterDomain == "" {
		opts.KubeClusterDomain = DefaultKubeClusterDomain
	}
	if opts.Mode == "" {
		opts.Mode = DefaultMode
	}
	return nil
}
