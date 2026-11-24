// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package permissions

import (
	"encoding/json"
	"fmt"
	"iter"
	"os"
)

// The types below model permissions_map.json: the output of tools/sdkmap, with
// IAM permissions added to every FuncRef. They mirror tools/sdkmap's output
// field for field and in the same order, so that sdkmap.json is written back
// unchanged apart from the added fields. Keep them in sync with tools/sdkmap.

// PermissionsMap is every resource and data source in the provider. It is
// read from sdkmap.json, and written as permissions_map.json once
// AddPermissions has filled in each FuncRef's IAM permissions.
type PermissionsMap struct {
	Resources   []Resource   `json:"resources"`
	DataSources []DataSource `json:"dataSources"`
}

type Resource struct {
	Service      string `json:"service"`
	ResourceType string `json:"resource_type"`
	PluginType   string `json:"plugin_type"`
	Expr         string `json:"expr"`
	Location     string `json:"location"`

	Create        *FuncRef      `json:"create,omitempty"`
	Read          *FuncRef      `json:"read,omitempty"`
	Update        *FuncRef      `json:"update,omitempty"`
	Delete        *FuncRef      `json:"delete,omitempty"`
	CustomizeDiff *FuncRef      `json:"customize_diff,omitempty"`
	ModifyPlan    *FuncRef      `json:"modify_plan,omitempty"`
	Import        *FuncRef      `json:"import,omitempty"`
	Tags          *ResourceTags `json:"tags,omitempty"`
}

type DataSource struct {
	Service        string `json:"service"`
	DataSourceType string `json:"data_source_type"`
	PluginType     string `json:"plugin_type"`
	Expr           string `json:"expr"`
	Location       string `json:"location"`

	Read *FuncRef      `json:"read,omitempty"`
	Tags *ResourceTags `json:"tags,omitempty"`
}

// FuncRef is a function implementing part of a resource or data source, and
// the AWS SDK calls reachable from it.
type FuncRef struct {
	Expr     string   `json:"expr"`
	Location string   `json:"location,omitempty"`
	SDKCalls []string `json:"sdk_calls"`

	// IAMPermissions is added by this tool: the sorted IAM actions required by
	// SDKCalls. It is omitted until set, then written even when empty.
	IAMPermissions []string `json:"iam_permissions,omitzero"`
	// UnresolvedSDKCalls is added by this tool: the SDKCalls that could not be
	// mapped to IAM actions, and why.
	UnresolvedSDKCalls []UnresolvedSDKCall `json:"unresolved_sdk_calls,omitempty"`

	Functions []*FuncRef `json:"functions,omitempty"`
}

type UnresolvedSDKCall struct {
	SDKCall string           `json:"sdk_call"`
	Reason  unresolvedReason `json:"reason"`
}

type ResourceTags struct {
	List   *FuncRef `json:"list,omitempty"`
	Update *FuncRef `json:"update,omitempty"`
}

// LoadSDKMap parses an sdkmap.json file.
func LoadSDKMap(path string) (*PermissionsMap, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m PermissionsMap
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &m, nil
}

// encode returns m formatted as sdkmap writes it.
func (m *PermissionsMap) encode() ([]byte, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// FuncRefs yields every function reference in m, including the functions
// nested in a CustomizeDiff.
func (m *PermissionsMap) FuncRefs() iter.Seq[*FuncRef] {
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
			if !walk(m.Resources[i].funcRefs()) {
				return
			}
		}
		for i := range m.DataSources {
			if !walk(m.DataSources[i].funcRefs()) {
				return
			}
		}
	}
}

// funcRefs returns the resource's top-level function references; any may be nil.
func (r *Resource) funcRefs() []*FuncRef {
	refs := []*FuncRef{r.Create, r.Read, r.Update, r.Delete, r.CustomizeDiff, r.ModifyPlan, r.Import}
	if r.Tags != nil {
		refs = append(refs, r.Tags.List, r.Tags.Update)
	}
	return refs
}

// funcRefs returns the data source's top-level function references; any may be nil.
func (d *DataSource) funcRefs() []*FuncRef {
	refs := []*FuncRef{d.Read}
	if d.Tags != nil {
		refs = append(refs, d.Tags.List, d.Tags.Update)
	}
	return refs
}
