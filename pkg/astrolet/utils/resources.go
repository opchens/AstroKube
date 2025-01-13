package utils

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
	AllowedPodNumber int
	ScalarResources  map[v1.ResourceName]int64
	score            int64
}

var unit v1.ResourceList
var cpuUnit int64
var memUnit int64

func init() {
	unit = v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1000m"),
		v1.ResourceMemory: resource.MustParse("1024Mi"),
	}
	cpuUnit = unit.Cpu().MilliValue()
	memUnit = unit.Memory().Value()
}

func NewResource(rl v1.ResourceList) *Resource {
	r := &Resource{
		ScalarResources: make(map[v1.ResourceName]int64),
	}
	r.Add(rl)
	return r
}

func (r *Resource) Equal(res *Resource) bool {
	if r == nil && res == nil {
		return true
	}
	if r == nil || res == nil {
		return false
	}
	if r.AllowedPodNumber != res.AllowedPodNumber || r.EphemeralStorage != res.EphemeralStorage || r.Memory != res.Memory || r.MilliCPU != res.MilliCPU {
		return false
	}
	if len(r.ScalarResources) != len(res.ScalarResources) {
		return false
	}
	for r, v := range r.ScalarResources {
		resv, ok := res.ScalarResources[r]
		if !ok || v != resv {
			return false
		}
	}
	return true
}

func (r *Resource) Score() {
	cpuScope := r.MilliCPU / cpuUnit
	memScope := r.Memory / memUnit
	if cpuScope < memScope {
		r.score = cpuScope
	} else {
		r.score = memScope
	}
}

func (r *Resource) Less(res *Resource) bool {
	if r.score == 0 {
		r.Score()
	}
	if res.score == 0 {
		res.Score()
	}
	return r.score < res.score
}

func (r *Resource) ToResourceList() v1.ResourceList {
	if r == nil {
		return nil
	}
	return v1.ResourceList{
		v1.ResourceCPU: *resource.NewMilliQuantity(
			r.MilliCPU, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(
			r.Memory, resource.BinarySI),
		v1.ResourcePods: *resource.NewQuantity(
			int64(r.AllowedPodNumber), resource.DecimalSI),
	}
}

func (r *Resource) Add(rl v1.ResourceList) {
	if r == nil {
		return
	}
	for rName, rQuant := range rl {
		switch rName {
		case v1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case v1.ResourceMemory:
			r.Memory += rQuant.Value()
		case v1.ResourcePods:
			r.AllowedPodNumber += int(rQuant.Value())
		default:
		}
	}
}

func (r *Resource) AddResource(res *Resource) {
	r.MilliCPU += res.MilliCPU
	r.Memory += res.Memory
	r.AllowedPodNumber += res.AllowedPodNumber
}

func (r *Resource) SubResource(res *Resource) {
	r.MilliCPU -= res.MilliCPU
	r.Memory -= res.Memory
	r.AllowedPodNumber -= res.AllowedPodNumber

}

func (r *Resource) Clone() *Resource {
	res := &Resource{
		MilliCPU:         r.MilliCPU,
		Memory:           r.Memory,
		AllowedPodNumber: r.AllowedPodNumber,
		EphemeralStorage: r.EphemeralStorage,
	}
	if r.ScalarResources != nil {
		res.ScalarResources = make(map[v1.ResourceName]int64)
		for k, v := range r.ScalarResources {
			res.ScalarResources[k] = v
		}
	}
	return res
}

func (r *Resource) Diff(res *Resource) *Resource {
	result := r.DeepCopy()
	result.SubResource(res)
	return result
}

func (r *Resource) DeepCopy() *Resource {
	return r.Clone()
}

func CalculateResource(pod *v1.Pod) (res Resource) {
	r := &res
	for _, c := range pod.Spec.Containers {
		r.Add(c.Resources.Requests)
	}
	r.AllowedPodNumber = 1
	return
}
