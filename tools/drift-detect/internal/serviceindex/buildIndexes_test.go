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
)

func TestBuildServiceIndexes(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	serviceRoot := filepath.Join(repoRoot, "internal", "service")
	writeTestFile(t, filepath.Join(serviceRoot, "example", "resource.go"), `package example

// @SDKResource("aws_example_resource", name="Example Resource")
// @FrameworkResource(aws_example_alias, name="Alias Resource")
func resource() {
	_ = Meta().S3Client()
	_ = Meta().EC2Client()
	_ = Meta().S3Client()
}`)
	writeTestFile(t, filepath.Join(serviceRoot, "example", "resource_test.go"), `package example
// @SDKResource("aws_ignored", name="Ignored")`)
	writeTestFile(t, filepath.Join(serviceRoot, "example", "notes.txt"), "@SDKResource(aws_ignored)")
	writeTestFile(t, filepath.Join(serviceRoot, "plain.go"), `package service
// no resource annotation`)

	got, err := BuildServiceIndexes(repoRoot)
	if err != nil {
		t.Fatalf("BuildServiceIndexes: %v", err)
	}

	wantExample := ResourceLocation{
		Resource:          "aws_example_resource",
		Service:           "example",
		File:              filepath.Join("example", "resource.go"),
		TFResourceSegment: "ExampleResource",
	}
	if locations := got.ResourceIndex["aws_example_resource"]; !reflect.DeepEqual(locations, []ResourceLocation{wantExample}) {
		t.Errorf("resource location = %#v, want %#v", locations, []ResourceLocation{wantExample})
	}
	if locations := got.ResourceIndex["aws_example_alias"]; len(locations) != 1 || locations[0].TFResourceSegment != "AliasResource" {
		t.Errorf("alias locations = %#v, want one AliasResource location", locations)
	}
	if _, ok := got.ResourceIndex["aws_ignored"]; ok {
		t.Error("test file annotation was included in ResourceIndex")
	}
	wantClients := []string{"EC2Client", "S3Client"}
	if clients := got.FileConnIndex[filepath.Join("example", "resource.go")]; !reflect.DeepEqual(clients, wantClients) {
		t.Errorf("clients = %#v, want %#v", clients, wantClients)
	}
}

func TestExtractClients(t *testing.T) {
	t.Parallel()

	got := extractClients([]byte(`package example
func f() {
	_ = AWSClient().S3Client()
	_ = Meta().EC2Client()
	_ = Meta().S3Client()
}`))
	want := []string{"EC2Client", "S3Client"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractClients() = %#v, want %#v", got, want)
	}
	if got := extractClients([]byte("package example")); got != nil {
		t.Errorf("extractClients without clients = %#v, want nil", got)
	}
}

func TestBuildAWSClientIndex(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	connsRoot := filepath.Join(repoRoot, "internal", "conns")
	writeTestFile(t, filepath.Join(connsRoot, "awsclient_gen.go"), `package conns
func (c *AWSClient) S3Client() *s3.Client { return nil }
func (c *AWSClient) EC2Client() *ec2.Client { return nil }`)
	writeTestFile(t, filepath.Join(connsRoot, "awsclient.go"), `package conns
func (c *AWSClient) S3Client() *other.Client { return nil }
func (c *AWSClient) IAMClient() *iam.Client { return nil }`)

	got, err := BuildAWSClientIndex(repoRoot)
	if err != nil {
		t.Fatalf("BuildAWSClientIndex: %v", err)
	}
	want := map[string]string{"S3Client": "s3", "EC2Client": "ec2", "IAMClient": "iam"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("client index = %#v, want %#v", got, want)
	}
}

func TestBuildAWSClientIndexMissingFile(t *testing.T) {
	t.Parallel()

	_, err := BuildAWSClientIndex(t.TempDir())
	if err == nil {
		t.Fatal("BuildAWSClientIndex() error = nil, want missing-file error")
	}
}

func TestFetchLatestDates(t *testing.T) {
	oldURL := githubTreeURL
	t.Cleanup(func() { githubTreeURL = oldURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Accept header = %q", r.Header.Get("Accept"))
		}
		_ = json.NewEncoder(w).Encode(githubTreeResponse{Tree: []githubTreeEntry{
			{Path: "models/s3/service/2006-03-01/s3-2006-03-01.json"},
			{Path: "models/s3/service/2006-02-01/old.json"},
			{Path: "models/app-auto-scaling/service/2016-02-06/app-auto-scaling-2016-02-06.json"},
			{Path: "models/not-a-model.json"},
			{Path: "models/s3/service/2006-03-01/readme.txt"},
		}})
	}))
	defer server.Close()
	githubTreeURL = server.URL

	got, err := fetchLatestDates()
	if err != nil {
		t.Fatalf("fetchLatestDates: %v", err)
	}
	want := map[string]string{"s3": "2006-03-01", "app-auto-scaling": "2016-02-06"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("models = %#v, want %#v", got, want)
	}
}

func TestFetchLatestDatesHTTPStatusesAndMalformedJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		code int
		body string
	}{
		{name: "not found", code: http.StatusNotFound, body: ""},
		{name: "server error", code: http.StatusInternalServerError, body: ""},
		{name: "malformed JSON", code: http.StatusOK, body: "{"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			oldURL := githubTreeURL
			githubTreeURL = server.URL
			_, err := fetchLatestDates()
			githubTreeURL = oldURL
			server.Close()
			if err == nil {
				t.Fatal("fetchLatestDates() error = nil, want error")
			}
		})
	}
}

func TestAWSServiceIndex(t *testing.T) {
	t.Parallel()

	idx := &AWSServiceIndex{
		models:  map[string]string{"s3": "2006-03-01", "app-auto-scaling": "2016-02-06"},
		aliases: map[string]string{"appautoscaling": "app-auto-scaling"},
	}
	if !idx.HasService("s3") || idx.HasService("missing") {
		t.Fatal("HasService returned an unexpected result")
	}
	path, namespace, err := idx.ResolveModelPath("app-auto-scaling")
	if err != nil {
		t.Fatalf("ResolveModelPath: %v", err)
	}
	if path != "models/app-auto-scaling/service/2016-02-06/app-auto-scaling-2016-02-06.json" || namespace != "com.amazonaws.appautoscaling" {
		t.Errorf("ResolveModelPath() = %q, %q", path, namespace)
	}
	if _, _, err := idx.ResolveModelPath("missing"); err == nil {
		t.Fatal("ResolveModelPath(missing) error = nil")
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
