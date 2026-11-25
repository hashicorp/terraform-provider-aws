// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package model defines permissions_map.json: each resource and data source,
// the AWS SDK calls its functions make (found by package sdkmap), and the IAM
// permissions they require (added by package permissions).
package model

import (
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"slices"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/internal/atomicfile"
)

// PluginType is the Terraform plugin library a resource is implemented with.
type PluginType string

const (
	PluginTypeSDKv2     PluginType = "SDKv2"
	PluginTypeFramework PluginType = "Framework"
)

// Map is every resource and data source found.
type Map struct {
	Resources   []Resource   `json:"resources"`
	DataSources []DataSource `json:"dataSources"`
}

// Resource is a resource and the functions that implement it. For a
// Terraform Plugin Framework resource, methods from embedded types declared in
// another package, e.g. framework.WithNoUpdate, are nil.
type Resource struct {
	Service      string     `json:"service"`
	ResourceType string     `json:"resource_type"`
	PluginType   PluginType `json:"plugin_type"`

	// Expr is the factory function annotated with the resource's annotation,
	// e.g. @SDKResource.
	Expr string `json:"expr"`
	// Location is where Expr is declared, as "<file>:<line>" relative to the
	// repository root.
	Location string `json:"location"`

	Create *FuncRef `json:"create,omitempty"`
	Read   *FuncRef `json:"read,omitempty"`
	Update *FuncRef `json:"update,omitempty"`
	Delete *FuncRef `json:"delete,omitempty"`

	// CustomizeDiff is SDKv2 only; nil if unset.
	CustomizeDiff *FuncRef `json:"customize_diff,omitempty"`

	// ModifyPlan is Framework only; nil if unset.
	ModifyPlan *FuncRef `json:"modify_plan,omitempty"`

	// Import is the custom import function (SDKv2) or ImportState method
	// (Framework); nil if absent or passthrough only.
	Import *FuncRef `json:"import,omitempty"`

	// Tags is nil unless the resource is annotated with
	// @Tags(identifierAttribute=...).
	Tags *ResourceTags `json:"tags,omitempty"`
}

// DataSource is a data source and the function that reads it.
type DataSource struct {
	Service        string     `json:"service"`
	DataSourceType string     `json:"data_source_type"`
	PluginType     PluginType `json:"plugin_type"`

	// Expr is the factory function annotated with @SDKDataSource or
	// @FrameworkDataSource.
	Expr string `json:"expr"`
	// Location is where Expr is declared, as "<file>:<line>" relative to the
	// repository root.
	Location string `json:"location"`

	Read *FuncRef `json:"read,omitempty"`

	// Tags is nil unless the data source is annotated with
	// @Tags(identifierAttribute=...).
	Tags *ResourceTags `json:"tags,omitempty"`
}

// FuncRef is a function or method of a resource, data source, or service
// package, and the AWS SDK calls reachable from it.
type FuncRef struct {
	// Expr is the reference as written, e.g. "resourceObjectCreate" or
	// "schema.NoopContext", or the method expression, e.g.
	// "(*objectResource).Create".
	Expr string `json:"expr"`

	// Location is the function's declaration, or else the function literal or
	// call expression, as "<file>:<line>" relative to the repository root. It
	// is empty if the function isn't declared in the package, e.g.
	// schema.NoopContext.
	Location string `json:"location,omitempty"`

	// SDKCalls is the sorted set of AWS SDK client methods reachable from the
	// function, e.g. "s3.PutObjectAcl".
	SDKCalls []string `json:"sdk_calls"`

	// IAMPermissions is the sorted IAM actions required by SDKCalls. It is
	// omitted until set, then written even when empty.
	IAMPermissions []string `json:"iam_permissions,omitzero"`
	// UnresolvedSDKCalls is the SDKCalls that could not be mapped to IAM
	// actions, and why.
	UnresolvedSDKCalls []UnresolvedSDKCall `json:"unresolved_sdk_calls,omitempty"`

	// Functions lists the functions combined by customdiff.All or
	// customdiff.Sequence, flattened.
	Functions []*FuncRef `json:"functions,omitempty"`

	// Warnings notes AWS SDK clients passed to unanalyzed functions, e.g. an
	// SDK waiter, whose calls are missing from SDKCalls. Logged, not written.
	Warnings []string `json:"-"`
}

// UnresolvedReason explains why an SDK call could not be mapped to IAM
// permissions.
type UnresolvedReason string

// UnresolvedSDKCall is an SDK call with no IAM mapping, and why.
type UnresolvedSDKCall struct {
	SDKCall string           `json:"sdk_call"`
	Reason  UnresolvedReason `json:"reason"`
}

// ResourceTags holds the service package methods through which the tags
// interceptor lists and updates a resource's or data source's tags.
type ResourceTags struct {
	// List is the service package's ListTags, called after functions that
	// don't set tags themselves.
	List *FuncRef `json:"list,omitempty"`

	// Update is the service package's UpdateTags, called during Update. It is
	// nil for a data source.
	Update *FuncRef `json:"update,omitempty"`
}

// Load parses a file written by Write.
func Load(path string) (*Map, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Map
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &m, nil
}

// Write atomically writes m to path.
func (m *Map) Write(path string) error {
	b, err := m.Encode()
	if err != nil {
		return err
	}

	return atomicfile.Write(path, b)
}

// Encode returns m as indented JSON.
func (m *Map) Encode() ([]byte, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// FuncRefs yields every function reference in m, including the functions
// nested in a CustomizeDiff.
func (m *Map) FuncRefs() iter.Seq[*FuncRef] {
	return func(yield func(*FuncRef) bool) {
		var walk func(refs []*FuncRef) bool
		walk = func(refs []*FuncRef) bool {
			for _, f := range refs {
				if f == nil {
					continue
				}
				if !yield(f) || !walk(f.Functions) {
					return false
				}
			}
			return true
		}

		for i := range m.Resources {
			if !walk(m.Resources[i].FuncRefs()) {
				return
			}
		}
		for i := range m.DataSources {
			if !walk(m.DataSources[i].FuncRefs()) {
				return
			}
		}
	}
}

// FuncRefs returns the resource's top-level function references that are
// set, including the tags interceptor's methods.
func (r *Resource) FuncRefs() []*FuncRef {
	refs := []*FuncRef{r.Create, r.Read, r.Update, r.Delete, r.CustomizeDiff, r.ModifyPlan, r.Import}
	if r.Tags != nil {
		refs = append(refs, r.Tags.List, r.Tags.Update)
	}
	return nonNil(refs)
}

// FuncRefs returns the data source's top-level function references that are
// set, including the tags interceptor's methods.
func (d *DataSource) FuncRefs() []*FuncRef {
	refs := []*FuncRef{d.Read}
	if d.Tags != nil {
		refs = append(refs, d.Tags.List, d.Tags.Update)
	}
	return nonNil(refs)
}

// nonNil returns refs without the nil ones.
func nonNil(refs []*FuncRef) []*FuncRef {
	return slices.DeleteFunc(refs, func(ref *FuncRef) bool { return ref == nil })
}
