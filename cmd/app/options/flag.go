package options

import "github.com/spf13/pflag"

func InstallFlags(flag *pflag.FlagSet, opts *ServerRunOptions) {
	flag.StringVar(&opts.ClusterName, "cluster-name", opts.ClusterName, "cluster name")
	flag.StringVar(&opts.CoreClusterKubeConfigPath, "corecluster-kubeconfig", opts.CoreClusterKubeConfigPath, "kubeconfig path to core cluster")
	flag.StringVar(&opts.SubClusterKubeConfigPath, "subcluster-kubeconfig", opts.SubClusterKubeConfigPath, "kubeconfig path to sub cluster")
	flag.StringVar(&opts.FedNamespacePrefix, "fed-namespace-prefix", opts.FedNamespacePrefix, "prefix of namespace, will add prefix to namespace while sync resources from star cluster to sub cluster")
	flag.StringVar(&opts.Mode, "mode", opts.Mode, "astrolet run mode")
}
