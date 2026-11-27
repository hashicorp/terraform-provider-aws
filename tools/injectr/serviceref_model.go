// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// ServiceRef is the parsed service reference for a single AWS service. Only
// the API operations, and the IAM actions each requires, are decoded.
type ServiceRef struct {
	Name       string      `json:"Name"` // IAM service prefix, e.g. "s3"
	Version    string      `json:"Version"`
	Operations []Operation `json:"Operations"`
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

// SDKMethod is an SDK client method implementing an operation, e.g. Boto3's
// "elbv2" client's "create_load_balancer".
type SDKMethod struct {
	Name    string `json:"Name"`
	Method  string `json:"Method"`
	Package string `json:"Package"`
}

// ServiceRefs maps an IAM service prefix (e.g. "s3") to its service reference.
type ServiceRefs map[string]*ServiceRef

// LoadServiceRef parses a service reference file written by GetServiceRef.
func LoadServiceRef(path string) (*ServiceRef, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ref, err := parseServiceRef(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return ref, nil
}

// parseServiceRef decodes a service reference, which must name its service.
func parseServiceRef(b []byte) (*ServiceRef, error) {
	var ref ServiceRef
	if err := json.Unmarshal(b, &ref); err != nil {
		return nil, err
	}
	if ref.Name == "" {
		return nil, errors.New("missing service name")
	}

	return &ref, nil
}

// LoadServiceRefs parses every <service>.json file in dir concurrently. Files
// that fail to parse are skipped and reported in the returned error.
func LoadServiceRefs(dir string) (ServiceRefs, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		refs = make(ServiceRefs, len(entries))
		errs []error
	)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		wg.Go(func() {
			ref, err := LoadServiceRef(filepath.Join(dir, e.Name()))

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			if _, ok := refs[ref.Name]; ok {
				errs = append(errs, fmt.Errorf("%s: duplicate service %q", e.Name(), ref.Name))
				return
			}
			refs[ref.Name] = ref
		})
	}
	wg.Wait()

	// Files finish in any order; sort so the reported errors are stable.
	slices.SortFunc(errs, func(a, b error) int { return cmp.Compare(a.Error(), b.Error()) })

	return refs, errors.Join(errs...)
}
