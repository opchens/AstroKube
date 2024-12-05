/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	corev1 "AstroKube/pkg/apis/core/v1"

	apicorev1 "k8s.io/api/core/v1"
)

// ClusterStatusApplyConfiguration represents an declarative configuration of the ClusterStatus type for use
// with apply.
type ClusterStatusApplyConfiguration struct {
	Addresses                 []AddressApplyConfiguration                   `json:"addresses,omitempty"`
	Phase                     *corev1.ClusterPhase                          `json:"phase,omitempty"`
	Allocatable               *apicorev1.ResourceList                       `json:"allocatable,omitempty"`
	Usage                     *apicorev1.ResourceList                       `json:"usage,omitempty"`
	Nodes                     []NodeLeftResourcesApplyConfiguration         `json:"nodes,omitempty"`
	SubClusterAllocatable     *apicorev1.ResourceList                       `json:"subClusterAllocatable,omitempty"`
	SubClusterUsage           *apicorev1.ResourceList                       `json:"subClusterUsage,omitempty"`
	SubClusterNodes           []NodeLeftResourcesApplyConfiguration         `json:"subClusterNodes,omitempty"`
	Namespace                 []NamespaceUsageApplyConfiguration            `json:"namespace,omitempty"`
	Condition                 []ClusterConditionApplyConfiguration          `json:"condition,omitempty"`
	ClusterInfo               *ClusterInfoApplyConfiguration                `json:"clusterInfo,omitempty"`
	DaemonEndpoint            *ClusterDaemonEndpointsApplyConfiguration     `json:"daemonEndpoint,omitempty"`
	Aggregate                 []string                                      `json:"aggregate,omitempty"`
	SecretRef                 *ClusterSecretRefApplyConfiguration           `json:"secretRef,omitempty"`
	Storage                   []string                                      `json:"storage,omitempty"`
	CustomResourceDefinitions []string                                      `json:"customResourceDefinitions,omitempty"`
	NodeAggregate             map[string]NodesInAggregateApplyConfiguration `json:"nodeAggregate,omitempty"`
}

// ClusterStatusApplyConfiguration constructs an declarative configuration of the ClusterStatus type for use with
// apply.
func ClusterStatus() *ClusterStatusApplyConfiguration {
	return &ClusterStatusApplyConfiguration{}
}

// WithAddresses adds the given value to the Addresses field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Addresses field.
func (b *ClusterStatusApplyConfiguration) WithAddresses(values ...*AddressApplyConfiguration) *ClusterStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithAddresses")
		}
		b.Addresses = append(b.Addresses, *values[i])
	}
	return b
}

// WithPhase sets the Phase field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Phase field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithPhase(value corev1.ClusterPhase) *ClusterStatusApplyConfiguration {
	b.Phase = &value
	return b
}

// WithAllocatable sets the Allocatable field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Allocatable field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithAllocatable(value apicorev1.ResourceList) *ClusterStatusApplyConfiguration {
	b.Allocatable = &value
	return b
}

// WithUsage sets the Usage field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Usage field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithUsage(value apicorev1.ResourceList) *ClusterStatusApplyConfiguration {
	b.Usage = &value
	return b
}

// WithNodes adds the given value to the Nodes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Nodes field.
func (b *ClusterStatusApplyConfiguration) WithNodes(values ...*NodeLeftResourcesApplyConfiguration) *ClusterStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithNodes")
		}
		b.Nodes = append(b.Nodes, *values[i])
	}
	return b
}

// WithSubClusterAllocatable sets the SubClusterAllocatable field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SubClusterAllocatable field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithSubClusterAllocatable(value apicorev1.ResourceList) *ClusterStatusApplyConfiguration {
	b.SubClusterAllocatable = &value
	return b
}

// WithSubClusterUsage sets the SubClusterUsage field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SubClusterUsage field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithSubClusterUsage(value apicorev1.ResourceList) *ClusterStatusApplyConfiguration {
	b.SubClusterUsage = &value
	return b
}

// WithSubClusterNodes adds the given value to the SubClusterNodes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the SubClusterNodes field.
func (b *ClusterStatusApplyConfiguration) WithSubClusterNodes(values ...*NodeLeftResourcesApplyConfiguration) *ClusterStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithSubClusterNodes")
		}
		b.SubClusterNodes = append(b.SubClusterNodes, *values[i])
	}
	return b
}

// WithNamespace adds the given value to the Namespace field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Namespace field.
func (b *ClusterStatusApplyConfiguration) WithNamespace(values ...*NamespaceUsageApplyConfiguration) *ClusterStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithNamespace")
		}
		b.Namespace = append(b.Namespace, *values[i])
	}
	return b
}

// WithCondition adds the given value to the Condition field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Condition field.
func (b *ClusterStatusApplyConfiguration) WithCondition(values ...*ClusterConditionApplyConfiguration) *ClusterStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithCondition")
		}
		b.Condition = append(b.Condition, *values[i])
	}
	return b
}

// WithClusterInfo sets the ClusterInfo field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ClusterInfo field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithClusterInfo(value *ClusterInfoApplyConfiguration) *ClusterStatusApplyConfiguration {
	b.ClusterInfo = value
	return b
}

// WithDaemonEndpoint sets the DaemonEndpoint field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DaemonEndpoint field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithDaemonEndpoint(value *ClusterDaemonEndpointsApplyConfiguration) *ClusterStatusApplyConfiguration {
	b.DaemonEndpoint = value
	return b
}

// WithAggregate adds the given value to the Aggregate field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Aggregate field.
func (b *ClusterStatusApplyConfiguration) WithAggregate(values ...string) *ClusterStatusApplyConfiguration {
	for i := range values {
		b.Aggregate = append(b.Aggregate, values[i])
	}
	return b
}

// WithSecretRef sets the SecretRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SecretRef field is set to the value of the last call.
func (b *ClusterStatusApplyConfiguration) WithSecretRef(value *ClusterSecretRefApplyConfiguration) *ClusterStatusApplyConfiguration {
	b.SecretRef = value
	return b
}

// WithStorage adds the given value to the Storage field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Storage field.
func (b *ClusterStatusApplyConfiguration) WithStorage(values ...string) *ClusterStatusApplyConfiguration {
	for i := range values {
		b.Storage = append(b.Storage, values[i])
	}
	return b
}

// WithCustomResourceDefinitions adds the given value to the CustomResourceDefinitions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the CustomResourceDefinitions field.
func (b *ClusterStatusApplyConfiguration) WithCustomResourceDefinitions(values ...string) *ClusterStatusApplyConfiguration {
	for i := range values {
		b.CustomResourceDefinitions = append(b.CustomResourceDefinitions, values[i])
	}
	return b
}

// WithNodeAggregate puts the entries into the NodeAggregate field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the NodeAggregate field,
// overwriting an existing map entries in NodeAggregate field with the same key.
func (b *ClusterStatusApplyConfiguration) WithNodeAggregate(entries map[string]NodesInAggregateApplyConfiguration) *ClusterStatusApplyConfiguration {
	if b.NodeAggregate == nil && len(entries) > 0 {
		b.NodeAggregate = make(map[string]NodesInAggregateApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.NodeAggregate[k] = v
	}
	return b
}