// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ServiceRef is the parsed service reference for a single AWS service: the
// IAM actions, resource types, and condition keys it defines, and the IAM
// actions each API operation requires.
type ServiceRef struct {
	Name          string         `json:"Name"`        // IAM service prefix, e.g. "s3"
	ServiceName   string         `json:"ServiceName"` // display name
	Version       string         `json:"Version"`
	Actions       []Action       `json:"Actions"`
	ConditionKeys []ConditionKey `json:"ConditionKeys"`
	Operations    []Operation    `json:"Operations"`
	Resources     []ResourceType `json:"Resources"`
}

// Action is an IAM action defined by a service, e.g. s3:GetObject.
type Action struct {
	Name                string            `json:"Name"`
	ActionConditionKeys []string          `json:"ActionConditionKeys"`
	Annotations         ActionAnnotations `json:"Annotations"`
	Resources           []ActionResource  `json:"Resources"`
	SupportedBy         map[string]bool   `json:"SupportedBy"`
}

type ActionAnnotations struct {
	Properties ActionProperties `json:"Properties"`
}

type ActionProperties struct {
	IsDataPlane            bool `json:"IsDataPlane"`
	IsList                 bool `json:"IsList"`
	IsPermissionManagement bool `json:"IsPermissionManagement"`
	IsTaggingOnly          bool `json:"IsTaggingOnly"`
	IsWrite                bool `json:"IsWrite"`
}

// ActionResource is a resource type an action can be scoped to.
type ActionResource struct {
	Name          string   `json:"Name"`
	ConditionKeys []string `json:"ConditionKeys"`
}

type ConditionKey struct {
	Name  string   `json:"Name"`
	Types []string `json:"Types"`
}

// Operation is an API operation, named as in the AWS SDKs (e.g.
// "PutBucketTagging"), and the IAM actions required to call it.
type Operation struct {
	Name              string             `json:"Name"`
	AuthorizedActions []AuthorizedAction `json:"AuthorizedActions"`
	SDK               []SDKMethod        `json:"SDK"`
}

// AuthorizedAction is an IAM action required by an operation. Service may
// differ from the operation's own service.
type AuthorizedAction struct {
	Name    string `json:"Name"`
	Service string `json:"Service"`
	// Context holds extra constraints on the action, e.g. the services a role
	// may be passed to under "iam:PassedToService".
	Context map[string][]string `json:"Context"`
}

// IAMAction returns the action in IAM policy form, e.g. "s3:GetObject".
func (a AuthorizedAction) IAMAction() string {
	return a.Service + ":" + a.Name
}

// SDKMethod is an SDK method implementing an operation. Only Boto3 is listed
// by the service reference, so match operations by Operation.Name instead.
type SDKMethod struct {
	Name    string `json:"Name"`
	Method  string `json:"Method"`
	Package string `json:"Package"`
}

// ResourceType is an IAM resource type defined by a service.
type ResourceType struct {
	Name          string   `json:"Name"`
	ARNFormats    []string `json:"ARNFormats"`
	ConditionKeys []string `json:"ConditionKeys"`
}

// Operation returns the operation with the given name.
func (s *ServiceRef) Operation(name string) (*Operation, bool) {
	return find(s.Operations, func(o Operation) bool { return o.Name == name })
}

// Action returns the IAM action with the given name, without the service prefix.
func (s *ServiceRef) Action(name string) (*Action, bool) {
	return find(s.Actions, func(a Action) bool { return a.Name == name })
}

// ResourceType returns the IAM resource type with the given name.
func (s *ServiceRef) ResourceType(name string) (*ResourceType, bool) {
	return find(s.Resources, func(r ResourceType) bool { return r.Name == name })
}

func find[T any](s []T, f func(T) bool) (*T, bool) {
	i := slices.IndexFunc(s, f)
	if i < 0 {
		return nil, false
	}
	return &s[i], true
}

// ServiceRefs maps an IAM service prefix (e.g. "s3") to its service reference.
type ServiceRefs map[string]*ServiceRef

// Operation returns the named operation of the named service.
func (r ServiceRefs) Operation(service, name string) (*Operation, bool) {
	ref, ok := r[service]
	if !ok {
		return nil, false
	}
	return ref.Operation(name)
}

// LoadServiceRef parses a service reference file written by GetServiceRef.
func LoadServiceRef(path string) (*ServiceRef, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ref ServiceRef
	if err := json.Unmarshal(b, &ref); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if ref.Name == "" {
		return nil, fmt.Errorf("%s: missing service name", path)
	}

	return &ref, nil
}

// LoadServiceRefs parses every <service>.json file in dir. Files that fail to
// parse are skipped and reported in the returned error.
func LoadServiceRefs(dir string) (ServiceRefs, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	refs := make(ServiceRefs, len(entries))
	var errs []error
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		ref, err := LoadServiceRef(filepath.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, ok := refs[ref.Name]; ok {
			errs = append(errs, fmt.Errorf("%s: duplicate service %q", e.Name(), ref.Name))
			continue
		}
		refs[ref.Name] = ref
	}

	return refs, errors.Join(errs...)
}
