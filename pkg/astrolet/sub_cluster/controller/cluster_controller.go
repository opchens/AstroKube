package controller

import (
	astrov1 "AstroKube/pkg/apis/astrocore/v1"
	"time"
)

type ClusterController struct {
	ClusterName        string
	Cluster            *astrov1.Cluster
	NamespacePrefix    string
	HeartbeatFrequency time.Duration

	PlanetApiServerProtocol string
	PlanetApiServerPort     int32

	clusterDaemonEndpoint astrov1.ClusterDaemonEndpoints
	apiServerAddress      []astrov1.Address
	clusterInfo           astrov1.ClusterInfo
	secretRef             astrov1.ClusterSecretRef

	clusterInformer
}
