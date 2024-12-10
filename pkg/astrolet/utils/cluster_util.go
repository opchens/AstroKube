package utils

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func UpdateClusterCondition(cluster *astrov1.Cluster, condition *astrov1.ClusterCondition) {
	if cluster == nil || condition == nil {
		return
	}
	index := -1
	traansition := false
	if cluster.Status.Condition == nil {
		cluster.Status.Condition = make([]astrov1.ClusterCondition, 1)
		index = 0
		traansition = true
	} else {
		for i, c := range cluster.Status.Condition {
			if c.Type == condition.Type {
				index = i
				if c.Status != condition.Status {
					traansition = true
				}
				break
			}
		}
	}
	if index == -1 {
		cluster.Status.Condition = append(cluster.Status.Condition, astrov1.ClusterCondition{})
		index = len(cluster.Status.Condition) - 1
		traansition = true
	}
	now := time.Now()
	cluster.Status.Condition[index].LastHeartbeatTime = v1.NewTime(now)
	if traansition {
		cluster.Status.Condition[index].LastTransitionTime = v1.NewTime(now)
	}
	cluster.Status.Condition[index].Type = condition.Type
	cluster.Status.Condition[index].Status = condition.Status
	cluster.Status.Condition[index].Reason = condition.Reason
	cluster.Status.Condition[index].Message = condition.Message
}
