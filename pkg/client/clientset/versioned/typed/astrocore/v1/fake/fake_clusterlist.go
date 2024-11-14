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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "AstroKube/pkg/apis/astrocore/v1"
	astrocorev1 "AstroKube/pkg/client/applyconfiguration/astrocore/v1"
	"context"
	json "encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterLists implements ClusterListInterface
type FakeClusterLists struct {
	Fake *FakeAstroV1
	ns   string
}

var clusterlistsResource = v1.SchemeGroupVersion.WithResource("clusterlists")

var clusterlistsKind = v1.SchemeGroupVersion.WithKind("ClusterList")

// Get takes name of the clusterList, and returns the corresponding clusterList object, and an error if there is any.
func (c *FakeClusterLists) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clusterlistsResource, c.ns, name), &v1.ClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterList), err
}

// List takes label and field selectors, and returns the list of ClusterLists that match those selectors.
func (c *FakeClusterLists) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ClusterListList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clusterlistsResource, clusterlistsKind, c.ns, opts), &v1.ClusterListList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterListList), err
}

// Watch returns a watch.Interface that watches the requested clusterLists.
func (c *FakeClusterLists) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clusterlistsResource, c.ns, opts))

}

// Create takes the representation of a clusterList and creates it.  Returns the server's representation of the clusterList, and an error, if there is any.
func (c *FakeClusterLists) Create(ctx context.Context, clusterList *v1.ClusterList, opts metav1.CreateOptions) (result *v1.ClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clusterlistsResource, c.ns, clusterList), &v1.ClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterList), err
}

// Update takes the representation of a clusterList and updates it. Returns the server's representation of the clusterList, and an error, if there is any.
func (c *FakeClusterLists) Update(ctx context.Context, clusterList *v1.ClusterList, opts metav1.UpdateOptions) (result *v1.ClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clusterlistsResource, c.ns, clusterList), &v1.ClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterList), err
}

// Delete takes name of the clusterList and deletes it. Returns an error if one occurs.
func (c *FakeClusterLists) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(clusterlistsResource, c.ns, name, opts), &v1.ClusterList{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterLists) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clusterlistsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ClusterListList{})
	return err
}

// Patch applies the patch and returns the patched clusterList.
func (c *FakeClusterLists) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterlistsResource, c.ns, name, pt, data, subresources...), &v1.ClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterList), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied clusterList.
func (c *FakeClusterLists) Apply(ctx context.Context, clusterList *astrocorev1.ClusterListApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ClusterList, err error) {
	if clusterList == nil {
		return nil, fmt.Errorf("clusterList provided to Apply must not be nil")
	}
	data, err := json.Marshal(clusterList)
	if err != nil {
		return nil, err
	}
	name := clusterList.Name
	if name == nil {
		return nil, fmt.Errorf("clusterList.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterlistsResource, c.ns, *name, types.ApplyPatchType, data), &v1.ClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ClusterList), err
}
