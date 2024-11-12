package sub_cluster

import "context"

type SubManager struct {
	ExternalIp  string
	ClusterName string
}

func (m SubManager) Run(ctx context.Context) {}
