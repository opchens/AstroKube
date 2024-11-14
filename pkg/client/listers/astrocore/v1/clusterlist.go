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
// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "AstroKube/pkg/apis/astrocore/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterListLister helps list ClusterLists.
// All objects returned here must be treated as read-only.
type ClusterListLister interface {
	// List lists all ClusterLists in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ClusterList, err error)
	// ClusterLists returns an object that can list and get ClusterLists.
	ClusterLists(namespace string) ClusterListNamespaceLister
	ClusterListListerExpansion
}

// clusterListLister implements the ClusterListLister interface.
type clusterListLister struct {
	indexer cache.Indexer
}

// NewClusterListLister returns a new ClusterListLister.
func NewClusterListLister(indexer cache.Indexer) ClusterListLister {
	return &clusterListLister{indexer: indexer}
}

// List lists all ClusterLists in the indexer.
func (s *clusterListLister) List(selector labels.Selector) (ret []*v1.ClusterList, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClusterList))
	})
	return ret, err
}

// ClusterLists returns an object that can list and get ClusterLists.
func (s *clusterListLister) ClusterLists(namespace string) ClusterListNamespaceLister {
	return clusterListNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClusterListNamespaceLister helps list and get ClusterLists.
// All objects returned here must be treated as read-only.
type ClusterListNamespaceLister interface {
	// List lists all ClusterLists in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ClusterList, err error)
	// Get retrieves the ClusterList from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.ClusterList, error)
	ClusterListNamespaceListerExpansion
}

// clusterListNamespaceLister implements the ClusterListNamespaceLister
// interface.
type clusterListNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClusterLists in the indexer for a given namespace.
func (s clusterListNamespaceLister) List(selector labels.Selector) (ret []*v1.ClusterList, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClusterList))
	})
	return ret, err
}

// Get retrieves the ClusterList from the indexer for a given namespace and name.
func (s clusterListNamespaceLister) Get(name string) (*v1.ClusterList, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("clusterlist"), name)
	}
	return obj.(*v1.ClusterList), nil
}
