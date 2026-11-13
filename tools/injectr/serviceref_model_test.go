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

func TestLoadServiceRefs(t *testing.T) {
	t.Parallel()

	refs, err := LoadServiceRefs("testdata/service-data")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := slices.Sorted(maps.Keys(refs)), []string{"iam", "s3"}; !slices.Equal(got, want) {
		t.Fatalf("services = %v, want %v", got, want)
	}

	s3 := refs["s3"]
	if s3.ServiceName != "S3" || s3.Version != "v1.5" {
		t.Errorf("s3 header = %q %q", s3.ServiceName, s3.Version)
	}

	a, ok := s3.Action("PutBucketTagging")
	if !ok {
		t.Fatal("action PutBucketTagging not found")
	}
	if !a.Annotations.Properties.IsTaggingOnly {
		t.Error("PutBucketTagging: IsTaggingOnly = false, want true")
	}
	if !a.SupportedBy["IAM Action Last Accessed"] {
		t.Error("PutBucketTagging: SupportedBy not parsed")
	}

	rt, ok := s3.ResourceType("bucket")
	if !ok || len(rt.ARNFormats) != 1 {
		t.Errorf("resource type bucket = %+v, %t", rt, ok)
	}

	if _, ok := s3.Operation("NoSuchOperation"); ok {
		t.Error("NoSuchOperation found")
	}
	if _, ok := refs.Operation("nosuchservice", "CopyObject"); ok {
		t.Error("nosuchservice found")
	}
}

func TestServiceRefsOperationIAMActions(t *testing.T) {
	t.Parallel()

	refs, err := LoadServiceRefs("testdata/service-data")
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		service, operation string
		want               []string
	}{
		"single action": {
			service:   "s3",
			operation: "PutBucketTagging",
			want:      []string{"s3:PutBucketTagging"},
		},
		"cross-service action": {
			service:   "s3",
			operation: "CopyObject",
			want:      []string{"s3:GetObject", "s3-object-lambda:PutObject"},
		},
		"no authorized actions": {
			service:   "s3",
			operation: "ListDirectoryBuckets",
		},
		"context": {
			service:   "iam",
			operation: "CreateInstanceProfile",
			want:      []string{"iam:PassRole"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			op, ok := refs.Operation(tc.service, tc.operation)
			if !ok {
				t.Fatalf("operation %s:%s not found", tc.service, tc.operation)
			}

			var got []string
			for _, aa := range op.AuthorizedActions {
				got = append(got, aa.IAMAction())
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("IAM actions = %v, want %v", got, tc.want)
			}
		})
	}

	op, _ := refs.Operation("iam", "CreateInstanceProfile")
	if got := op.AuthorizedActions[0].Context["iam:PassedToService"]; !slices.Equal(got, []string{"ec2.amazonaws.com"}) {
		t.Errorf("Context[iam:PassedToService] = %v", got)
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
	if _, ok := refs["good"]; !ok {
		t.Error("good not loaded")
	}
	if _, ok := refs["dupe"]; !ok {
		t.Error("first dupe not loaded")
	}
	if len(refs) != 2 {
		t.Errorf("loaded %d services, want 2", len(refs))
	}

	if _, err := LoadServiceRefs(filepath.Join(dir, "missing")); err == nil {
		t.Error("expected error for missing dir")
	}
}
