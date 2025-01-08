package common

var (
	AstroSystem = "astro-system"

	IgnoredCustomResourcesConfigMap = "ignored-custom-resources"
	IgnoredCustomResourcesData      = "ignored-custom-resources"

	IgnoredCustomResources = []string{
		"monitoring.coreos.com/v1alpha1, Kind=AlertmanagerConfig",
		"monitoring.coreos.com/v1, Kind=AlertManager",
		"monitoring.coreos.com/v1, Kind=PodMonitor",
		"monitoring.coreos.com/v1, Kind=Probe",
		"monitoring.coreos.com/v1, Kind=Prometheus",
		"monitoring.coreos.com/v1, Kind=PrometheusRule",
		"monitoring.coreos.com/v1, Kind=ServiceMonitor",
		"monitoring.coreos.com/v1, Kind=ThanosRuler",
	}
)
