package app

import (
	"AstroKube/cmd/app/options"
	"AstroKube/pkg/astrolet"
	"AstroKube/pkg/astrolet/utils"
	"AstroKube/pkg/client/clientset/versioned"
	"context"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"
	"os"
)

const componentAstroLet = "AstroLet"

func NewAstroLetCommand(ctx context.Context, opts options.ServerRunOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: componentAstroLet,
		Long: `The AstroLet is the primary "cluster agent" that runs on each
SubCluster. It can register the cluster with the CoreCluster apiserver.

The AstroLet works in terms of a PodSpec or a WorkloadSpec. A PodSpec or a WorkloadSpec is a YAML or JSON object
that describes a pod or a workload. The AstroLet ensures that the pods
described in those PodSpecs and WorkloadSpec are running and healthy. The kubelet doesn't manage
pods which were not created by CoreCluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
			return runCmd(ctx, opts)
		},
	}
	verflag.AddFlags(cmd.Flags())
	options.InstallFlags(cmd.Flags(), &opts)
	return cmd
}

func runCmd(ctx context.Context, opts options.ServerRunOptions) error {
	if os.Getenv("KUBECONFIG") == "" {
		os.Setenv("KUBECONFIG", opts.SubClusterKubeConfigPath)
	}
	os.Setenv("CORECLUSTERKUBECONFIG", opts.CoreClusterKubeConfigPath)
	klog.InfoS("Starting AstroLet ...", "LeaderElect", opts.LeaderElect)

	subClientConfig, subClientSet, subAstClientSet, err := initializeSubClusterClient(ctx, opts.SubClusterKubeConfigPath, opts.KubeApiQPSForSub, opts.KubeApiBurstForSub)
	if err != nil {
		klog.Errorf("cannot initialize clientset for  sub cluster: %v", err)
		return err
	}

	if !opts.LeaderElect {
		run(ctx, opts, subClientConfig, subClientSet, subAstClientSet)
		panic("unreachable")
	}
	id, err := os.Hostname()
	if err != nil {
		return err
	}

	// add an uniquifier so that two processes on the same host don't accidentally both become active
	id = id + "_" + string(uuid.NewUUID())

	eventRecorder := createRecorder(subClientSet, componentAstroLet)
	rl, err := resourcelock.New(opts.LeaderElectResourceLock,
		opts.LeaderElectResourceNamespace,
		opts.LeaderElectResourceName,
		subClientSet.CoreV1(),
		subClientSet.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		klog.Fatalf("error creating lock: %v", err)
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: opts.LeaderElectLeaseDuration,
		RenewDeadline: opts.LeaderElectRenewDeadline,
		RetryPeriod:   opts.LeaderElectRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				run(ctx, opts, subClientConfig, subClientSet, subAstClientSet)
			},
			OnStoppedLeading: func() {
				klog.Fatalf("leader election lost")
			},
			OnNewLeader: func(identity string) {
				// be notified when new leader elected
				if identity == id {
					// just got the lock
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
		Name: componentAstroLet,
	})
	panic("unreachable")
}

func run(ctx context.Context, opts options.ServerRunOptions, subClientConfig *rest.Config,
	subClientSet kubernetes.Interface, subAstClientSet versioned.Interface) error {
	ctx = utils.ContextInit(ctx)
	klog.Info("initialize core cluster client ...")
	coreConfig, coreClientSet, coreAstClientSet, err := initializeCoreClusterClient(ctx, opts.CoreClusterKubeConfigPath, opts.KubeApiQPSForCore, opts.KubeApiBurstForCore)
	if err != nil {
		klog.Errorf("cannot initialize clientset for core cluster: %v", err)
		return err
	}
	astrolet := NewAstroKet(coreConfig, coreClientSet, subClientConfig, subClientSet, subAstClientSet, coreAstClientSet, opts)
	klog.Info(" astrolet running ... ")
	astrolet.Run(ctx)

	utils.ContextStart(ctx)

	utils.ContextShutdown(ctx)
	return nil
}

func initializeSubClusterClient(ctx context.Context, kubeconfig string, kubeApiOPS float32, kubeApiBurst int) (*rest.Config, kubernetes.Interface, versioned.Interface, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Warning("inclusterConfig error, build config from flags ...")
		config, err = clientcmd.BuildConfigFromFlags("", "/Users/apple/.kube/config")
		if err != nil {
			klog.Errorf("cannot load config for sub cluster: %v", err)
			return nil, nil, nil, err
		}
	}

	config.QPS = kubeApiOPS
	config.Burst = kubeApiBurst

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	asclientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	return config, clientSet, asclientSet, nil
}

func initializeCoreClusterClient(ctx context.Context, kubeconfig string, kubeApiOps float32, kubeApiBurst int) (*rest.Config, kubernetes.Interface, versioned.Interface, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/apple/.kube/core.config")
	if err != nil {
		klog.Errorf("cannot load config for core cluster: %v", err)
	}

	config.QPS = kubeApiOps
	config.Burst = kubeApiBurst

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	astclientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	return config, clientSet, astclientSet, nil
}

func createRecorder(kubeClient kubernetes.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

func NewAstroKet(coreConfig *rest.Config, coreClientSet kubernetes.Interface,
	subClientConfig *rest.Config, subClientSet kubernetes.Interface, subAstClientSet, coreAstClientSet versioned.Interface,
	opts options.ServerRunOptions) *astrolet.AstroLet {
	// init AstroLet
	return &astrolet.AstroLet{
		CoreClient: coreClientSet,
		CoreConfig: coreConfig,
		SubClient:  subClientSet,
		SubConfig:  subClientConfig,

		SubAsClient:  subAstClientSet,
		CoreAsClient: coreAstClientSet,

		ClusterName:     opts.ClusterName,
		SubExternalIP:   opts.SubClusterExternalIp,
		NamespacePrefix: opts.FedNamespacePrefix,
		Mode:            opts.Mode,

		ForceSyncFrequency:   opts.ForceSyncFrequency,
		InformerResyncPeriod: opts.InformerResyncPeriod,
	}

}
