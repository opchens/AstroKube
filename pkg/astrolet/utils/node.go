package utils

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	"strconv"
)

func GetSubClusterNodeLevel(node *v1.Node) int {
	if node == nil || node.Labels == nil {
		return -1
	}
	level, err := strconv.Atoi(node.Labels[astrov1.SubClusterNode])
	if err != nil {
		return -1
	}
	return level
}
