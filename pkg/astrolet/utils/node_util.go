package utils

import (
	astrov1 "AstroKube/pkg/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"strings"
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

func IsNodeReady(node *v1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == v1.NodeReady {
			return c.Status == v1.ConditionTrue
		}
	}
	return false
}

type PatchStatus struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type NodeInfo struct {
	Node *v1.Node

	Allocatable *Resource
	Usage       *Resource

	Pods map[string]Resource

	Index int
	score int64
}

func NewNodeInfo(node *v1.Node) *NodeInfo {
	return &NodeInfo{
		Node:        node,
		Usage:       NewResource(nil),
		Allocatable: NewResource(node.Status.Allocatable),
		Pods:        make(map[string]Resource),
	}
}

func (n *NodeInfo) AddPod(pod *v1.Pod) *Resource {
	r := CalculateResource(pod)
	n.Usage.AddResource(&r)
	return &r
}

func (n *NodeInfo) RemovePod(pod *v1.Pod) *Resource {
	r := CalculateResource(pod)
	n.Usage.SubResource(&r)
	key, _ := cache.MetaNamespaceKeyFunc(pod)
	delete(n.Pods, key)
	return &r
}

func (n *NodeInfo) Score() {
	cpuScope := (n.Allocatable.MilliCPU - n.Usage.MilliCPU) / cpuUnit
	memScope := (n.Allocatable.Memory - n.Usage.Memory) / memUnit
	if cpuScope < memScope {
		n.score = cpuScope
	} else {
		n.score = memScope
	}
}

type ScoreNodeList []*NodeInfo

func (N ScoreNodeList) Len() int {
	return len(N)
}

func (N ScoreNodeList) Less(i, j int) bool {
	return N[i].score > N[j].score || (N[i].score == N[j].score && strings.Compare(N[i].Node.Name, N[j].Node.Name) == -1)
}

func (N ScoreNodeList) Swap(i, j int) {
	N[i].Index, N[j].Index = N[j].Index, N[i].Index
	N[i], N[j] = N[j], N[i]
}

func (N ScoreNodeList) Up(i int) (changed bool) {
	for i > 0 && N.Less(i, i-1) {
		changed = true
		N.Swap(i, i-1)
		i--
	}
	return
}

func (N ScoreNodeList) Down(i int) (changed bool) {
	n := N.Len() - 1
	for i < n && N.Less(i+1, i) {
		changed = true
		N.Swap(i, i+1)
		i++
	}
	return
}

func (N *ScoreNodeList) Remove(i int) {
	for i < N.Len()-1 {
		(*N)[i] = (*N)[i+1]
		(*N)[i].Index = i
		i++
	}
	(*N) = (*N)[:N.Len()-1]
}

func (N ScoreNodeList) Top(n int) []astrov1.NodeLeftResource {
	if N.Len() < n {
		n = N.Len()
	}
	ret := make([]astrov1.NodeLeftResource, 0, n)
	for i := 0; i < n; i++ {
		left := N[i].Allocatable.DeepCopy()
		left.SubResource(N[i].Usage)
		ret = append(ret, astrov1.NodeLeftResource{
			Name: N[i].Node.Name,
			Left: left.ToResourceList(),
		})
	}
	return ret
}
