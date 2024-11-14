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
// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	astrocorev1 "AstroKube/pkg/apis/astrocore/v1"
	versioned "AstroKube/pkg/client/clientset/versioned"
	internalinterfaces "AstroKube/pkg/client/informers/externalversions/internalinterfaces"
	v1 "AstroKube/pkg/client/listers/astrocore/v1"
	"context"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// WorkloadInformer provides access to a shared informer and lister for
// Workloads.
type WorkloadInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.WorkloadLister
}

type workloadInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWorkloadInformer constructs a new informer for Workload type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWorkloadInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWorkloadInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWorkloadInformer constructs a new informer for Workload type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWorkloadInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AstroV1().Workloads(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AstroV1().Workloads(namespace).Watch(context.TODO(), options)
			},
		},
		&astrocorev1.Workload{},
		resyncPeriod,
		indexers,
	)
}

func (f *workloadInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWorkloadInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *workloadInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&astrocorev1.Workload{}, f.defaultInformer)
}

func (f *workloadInformer) Lister() v1.WorkloadLister {
	return v1.NewWorkloadLister(f.Informer().GetIndexer())
}