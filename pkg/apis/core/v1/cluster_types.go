/*
Copyright 2024.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct{}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	Addresses []Address    `json:"addresses,omitempty"`
	Phase     ClusterPhase `json:"phase,omitempty"`

	// Allocatable indicate the resources available for scheduling  within the cluster
	Allocatable corev1.ResourceList `json:"allocatable,omitempty"`

	// Usage indicate the resources that has used for the cluster
	Usage corev1.ResourceList `json:"usage,omitempty"`

	// Nodes describes remained resource of top N nodes based on remaining resources
	Nodes []NodeLeftResources `json:"nodes,omitempty"`

	// SubClusterAllocated indicate the resources in the cluster that can be used for scheduling
	SubClusterAllocatable corev1.ResourceList `json:"subClusterAllocatable,omitempty"`

	//SubClusterUsage indicate the resources used by clusters
	SubClusterUsage corev1.ResourceList `json:"subClusterUsage,omitempty"`

	// SubClusterNodes indicted the resource of topN nodes based on remaining resources
	SubClusterNodes []NodeLeftResources `json:"subClusterNodes,omitempty"`

	// Namespace indicated the resource occupation of a federal namespace in the workloadCluster
	Namespace []NamespaceUsage `json:"namespace,omitempty"`

	// Conditions is an array of cluster conditions
	Condition []ClusterCondition `json:"condition,omitempty"`

	// ClusterInfo
	ClusterInfo ClusterInfo `json:"clusterInfo,omitempty"`

	//Endpoints of daemons running on the Cluster
	DaemonEndpoint ClusterDaemonEndpoints `json:"daemonEndpoint,omitempty"`

	// Aggregate indicate which aggregate cluster contains
	Aggregate []string `json:"aggregate,omitempty"`

	// SecretRef indicate the secretRef of sub cluster
	SecretRef ClusterSecretRef `json:"secretRef,omitempty"`

	// Storage is an array of csi plugins installed in the subCluster
	Storage []string `json:"storage,omitempty"`

	// CustomResourceDefinitions
	CustomResourceDefinitions []string `json:"customResourceDefinitions,omitempty"`

	// NodeAggregate describe remained resource of top N nodes in same partition based on remaining resources
	NodeAggregate map[string]NodesInAggregate `json:"nodeAggregate,omitempty"`
}

type NodesInAggregate struct {
	Nodes []NodeLeftResources `json:"nodes"`
}

type Address struct {
	AddressIP string      `json:"addressIP,omitempty"`
	Type      AddressType `json:"type,omitempty"`
}

type AddressType string

const (
	InternalIP AddressType = "InternalIP"
	ExternalIP AddressType = "ExternalIP"
)

type ClusterPhase string

const (
	ONLINE     ClusterPhase = "Online"
	OFFLINE    ClusterPhase = "Offline"
	PENDING    ClusterPhase = "Pending"
	TERMINATED ClusterPhase = "Terminated"
)

// NodeLeftResources Describes the remaining resources of the node
type NodeLeftResources struct {
	// Name  indicate the name of node
	Name string `json:"name"`

	Left corev1.ResourceList `json:"left"`
}

// NamespaceUsage Namespace Usage describe requests and limits resource of a federal namespace in the WorkloadCluster
type NamespaceUsage struct {
	Name string `json:"name"`

	// Usage indicate the requests and limits resource of a federal namespaces in the WorkloadCluster
	Usage corev1.ResourceRequirements `json:"usage"`
}

type ConditionStatus string

const (
	TrueCondition    = "True"
	FalseCondition   = "False"
	UnknownCondition = "Unknown"
)

type ClusterCondition struct {
	// Type of cluster  condition
	Type string `json:"type"`

	// Status indicate the status of the condition
	Status ConditionStatus `json:"status"`

	// LastHeartbeatTime indicate Last time we got an update on a given condition.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime"`

	// LastTransitionTime indicate  the condition transit from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// Reason indicate the reason of the condition's  last transtion
	Reason string `json:"reason"`

	// Message indicate details about last transtion
	Message string `json:"message"`
}

type ClusterInfo struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// DaemonEndpoint defined information about a single Daemon endpoint
type DaemonEndpoint struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
}

type ClusterDaemonEndpoints struct {

	// AstroletEndpoint indicate Endpoint on which astrolet is listening
	AstroletEndpoint DaemonEndpoint `json:"astroletEndpoint"`

	// ApiServerEndpoint indicate Endpoint on which KubeApiserver is listening in Sub Cluster
	ApiServerEndpoint DaemonEndpoint `json:"apiServerEndpoint"`
}

type ClusterSecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Platform",type="string",priority=0,JSONPath=".Status.ClusterInfo.Platform",description="The cluster status"
// +kubebuilder:printcolumn:name="Status",type="string",priority=0,JSONPath=".Status.Phase",description="The cluster status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

const (
	ClusterLevelLabel       = "astro.opchens.io/cluster-level"
	ClusterLevelCoreCluster = "1"
	ClusterLevelSubCluster  = "2"
)
