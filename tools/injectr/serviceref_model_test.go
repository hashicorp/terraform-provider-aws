// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const (
	testServiceDataDir = "testdata/service-data"
	realServiceDataDir = "service-data"
)

func TestLoadServiceRefs(t *testing.T) {
	t.Parallel()

	refs, err := LoadServiceRefs(testServiceDataDir)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := slices.Sorted(maps.Keys(refs)), []string{"elasticloadbalancing", "iam", "qldb", "s3"}; !slices.Equal(got, want) {
		t.Fatalf("services = %v, want %v", got, want)
	}

	s3 := refs["s3"]
	if s3.Version != "v1.5" {
		t.Errorf("s3 version = %q", s3.Version)
	}
	if got := len(s3.Operations); got != 3 {
		t.Fatalf("s3 operations = %d, want 3", got)
	}

	copyObject := s3.Operations[0]
	var got []string
	for _, aa := range copyObject.AuthorizedActions {
		got = append(got, aa.IAMAction())
	}
	if want := []string{"s3:GetObject", "s3-object-lambda:PutObject"}; !slices.Equal(got, want) {
		t.Errorf("CopyObject IAM actions = %v, want %v", got, want)
	}
	if want := (SDKMethod{Name: "s3", Method: "copy_object", Package: "Boto3"}); !slices.Equal(copyObject.SDK, []SDKMethod{want}) {
		t.Errorf("CopyObject SDK = %v, want %v", copyObject.SDK, want)
	}

	iam := refs["iam"].Operations[0].AuthorizedActions[0]
	if got := iam.Context["iam:PassedToService"]; !slices.Equal(got, []string{"ec2.amazonaws.com"}) {
		t.Errorf("Context[iam:PassedToService] = %v", got)
	}

	if got := len(refs["qldb"].Operations); got != 0 {
		t.Errorf("qldb operations = %d, want 0", got)
	}
}

func TestLoadServiceRefsErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := map[string]string{
		"good.json":   `{"Name": "good"}`,
		"bad.json":    `{`,
		"noname.json": `{}`,
		"ignored.txt": `not json`,
		"dupe-a.json": `{"Name": "dupe"}`,
		"dupe-b.json": `{"Name": "dupe"}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	refs, err := LoadServiceRefs(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	// Valid files still load alongside the errors.
	if got, want := slices.Sorted(maps.Keys(refs)), []string{"dupe", "good"}; !slices.Equal(got, want) {
		t.Errorf("services = %v, want %v", got, want)
	}

	if _, err := LoadServiceRefs(filepath.Join(dir, "missing")); err == nil {
		t.Error("expected error for missing dir")
	}
}

func BenchmarkLoadServiceRefs(b *testing.B) {
	skipWithoutRealData(b)

	for b.Loop() {
		if _, err := LoadServiceRefs(realServiceDataDir); err != nil {
			b.Fatal(err)
		}
	}
}

// skipWithoutRealData skips tb unless the tool has been run, so that the real
// service-data and sdkmap.json exist.
func skipWithoutRealData(tb testing.TB) {
	tb.Helper()

	for _, path := range []string{realServiceDataDir, realSDKMapFile} {
		if _, err := os.Stat(path); err != nil {
			tb.Skipf("%s not found; run the tool first", path)
		}
	}
}
