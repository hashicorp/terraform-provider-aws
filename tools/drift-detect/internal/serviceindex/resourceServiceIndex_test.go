// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package serviceindex

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

func TestBuildResourceServiceIndex(t *testing.T) {
	repoRoot := t.TempDir()
	serviceRoot := filepath.Join(repoRoot, "internal", "service")
	writeTestFile(t, filepath.Join(serviceRoot, "example", "resource.go"), `package example
// @SDKResource("aws_example_resource", name="Example Resource")
func resource() { _ = Meta().S3Client() }`)
	writeTestFile(t, filepath.Join(serviceRoot, "appautoscaling", "resource.go"), `package appautoscaling
// @FrameworkResource(aws_alias_resource, name="Alias Resource")
func resource() { _ = Meta().AppautoscalingClient() }`)
	writeTestFile(t, filepath.Join(serviceRoot, "multi", "resource.go"), `package multi
// @SDKResource("aws_multi_resource", name="Multi Resource")
func resource() { _ = Meta().EC2Client(); _ = Meta().S3Client() }`)
	writeTestFile(t, filepath.Join(repoRoot, "internal", "conns", "awsclient_gen.go"), `package conns
func (c *AWSClient) S3Client() *s3.Client { return nil }
func (c *AWSClient) AppautoscalingClient() *appautoscaling.Client { return nil }`)
	writeTestFile(t, filepath.Join(repoRoot, "internal", "conns", "awsclient.go"), `package conns
func (c *AWSClient) EC2Client() *ec2.Client { return nil }`)

	server := newTreeServer(t, []string{
		"models/s3/service/2006-03-01/s3-2006-03-01.json",
		"models/app-auto-scaling/service/2016-02-06/app-auto-scaling-2016-02-06.json",
	})
	defer server.Close()
	oldURL := githubTreeURL
	githubTreeURL = server.URL
	t.Cleanup(func() { githubTreeURL = oldURL })

	schema := &tfschema.ProviderSchema{Resources: map[string]*tfschema.ResourceIR{
		"aws_alias_resource":   {Name: "aws_alias_resource"},
		"aws_example_resource": {Name: "aws_example_resource"},
		"aws_missing_resource": {Name: "aws_missing_resource"},
		"aws_multi_resource":   {Name: "aws_multi_resource"},
		"aws_skipped_resource": {Name: "aws_skipped_resource"},
	}}
	skip := map[string]*awsmapping.SkipResource{
		"aws_skipped_resource": {Reason: "not supported by the drift detector"},
	}

	got, err := BuildResourceServiceIndex(repoRoot, schema, skip)
	if err != nil {
		t.Fatalf("BuildResourceServiceIndex: %v", err)
	}

	if got["aws_skipped_resource"].Skipped == nil || got["aws_skipped_resource"].Skipped.Reason != skip["aws_skipped_resource"].Reason {
		t.Errorf("skipped resource = %#v", got["aws_skipped_resource"])
	}
	assertInfo(t, got["aws_example_resource"], ResourceServiceInfo{
		Resource: "aws_example_resource", TFFile: filepath.Join("example", "resource.go"), TFResourceSegment: "ExampleResource",
		AWSClient: "S3Client", AWSService: "s3", AWSFile: "models/s3/service/2006-03-01/s3-2006-03-01.json", AWSNamespace: "com.amazonaws.s3",
	})
	assertInfo(t, got["aws_alias_resource"], ResourceServiceInfo{
		Resource: "aws_alias_resource", TFFile: filepath.Join("appautoscaling", "resource.go"), TFResourceSegment: "AliasResource",
		AWSClient: "AppautoscalingClient", AWSService: "app-auto-scaling", AWSFile: "models/app-auto-scaling/service/2016-02-06/app-auto-scaling-2016-02-06.json", AWSNamespace: "com.amazonaws.appautoscaling",
	})
	if got["aws_missing_resource"].Error != "Could not find TF file defining resource 'aws_missing_resource'" {
		t.Errorf("missing resource error = %q", got["aws_missing_resource"].Error)
	}
	if got["aws_multi_resource"].Error == "" {
		t.Error("multiple-client resource has no error")
	}
}

func TestResourceServiceCache(t *testing.T) {
	t.Parallel()

	cachePath := filepath.Join(t.TempDir(), ".cache", cacheFileName)
	want := map[string]*ResourceServiceInfo{
		"aws_example": {Resource: "aws_example", AWSService: "s3", Error: "warning"},
	}
	if err := writeResourceServiceCache(cachePath, want); err != nil {
		t.Fatalf("writeResourceServiceCache: %v", err)
	}
	if !isResourceServiceCacheFresh(cachePath) {
		t.Fatal("new cache is not fresh")
	}
	got, err := readResourceServiceCache(cachePath)
	if err != nil {
		t.Fatalf("readResourceServiceCache: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cache = %#v, want %#v", got, want)
	}
	got, err = ReadResourceServiceCache(cachePath)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ReadResourceServiceCache() = %#v, %v; want %#v, nil", got, err, want)
	}
}

func TestReadResourceServiceCacheErrors(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "missing.json")
	if _, err := readResourceServiceCache(missing); err == nil {
		t.Fatal("missing cache error = nil")
	}
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadResourceServiceCache(path); err == nil {
		t.Fatal("invalid cache error = nil")
	}
}

func TestIsResourceServiceCacheFreshExpired(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cache.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-(cacheTTL + time.Minute))
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	if isResourceServiceCacheFresh(path) {
		t.Fatal("expired cache reported fresh")
	}
}

func TestLoadResourceServiceIndexRefresh(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	cachePath := filepath.Join(repoRoot, ".cache", cacheFileName)
	want := map[string]*ResourceServiceInfo{"aws_cached": {Resource: "aws_cached"}}
	if err := writeResourceServiceCache(cachePath, want); err != nil {
		t.Fatal(err)
	}
	schema := &tfschema.ProviderSchema{Resources: map[string]*tfschema.ResourceIR{"aws_cached": {Name: "aws_cached"}}}
	mapping := &awsmapping.File{SkipResources: map[string]*awsmapping.SkipResource{}}
	got, err := LoadResourceServiceIndex(repoRoot, false, schema, mapping)
	if err != nil {
		t.Fatalf("LoadResourceServiceIndex: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("loaded cache = %#v, want %#v", got, want)
	}
}

func assertInfo(t *testing.T, got *ResourceServiceInfo, want ResourceServiceInfo) {
	t.Helper()
	if got == nil {
		t.Fatalf("resource info is nil, want %#v", want)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("resource info = %#v, want %#v", *got, want)
	}
}

func newTreeServer(t *testing.T, paths []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tree := make([]githubTreeEntry, 0, len(paths))
		for _, path := range paths {
			tree = append(tree, githubTreeEntry{Path: path, Type: "blob"})
		}
		if err := json.NewEncoder(w).Encode(githubTreeResponse{Tree: tree}); err != nil {
			t.Errorf("encode tree response: %v", err)
		}
	}))
}
