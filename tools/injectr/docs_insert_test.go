// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSelectDocs(t *testing.T) {
	t.Parallel()

	docs := []doc{
		{path: "website/docs/r/s3_bucket.html.markdown", service: "s3"},
		{path: "website/docs/d/s3_bucket.html.markdown", service: "s3"},
		{path: "website/docs/r/sqs_queue.html.markdown", service: "sqs"},
	}

	testCases := map[string]struct {
		opts    InsertOptions
		want    []string
		wantErr bool
	}{
		"all": {
			opts: InsertOptions{All: true},
			want: []string{"website/docs/r/s3_bucket.html.markdown", "website/docs/d/s3_bucket.html.markdown", "website/docs/r/sqs_queue.html.markdown"},
		},
		"file": {
			opts: InsertOptions{Files: []string{"./website/docs/d/s3_bucket.html.markdown"}},
			want: []string{"website/docs/d/s3_bucket.html.markdown"},
		},
		"service": {
			opts: InsertOptions{Services: []string{"s3"}},
			want: []string{"website/docs/r/s3_bucket.html.markdown", "website/docs/d/s3_bucket.html.markdown"},
		},
		"file and service overlap": {
			opts: InsertOptions{Files: []string{"website/docs/r/s3_bucket.html.markdown"}, Services: []string{"s3"}},
			want: []string{"website/docs/r/s3_bucket.html.markdown", "website/docs/d/s3_bucket.html.markdown"},
		},
		"unknown file": {
			opts:    InsertOptions{Files: []string{"website/docs/r/nope.html.markdown"}},
			wantErr: true,
		},
		"unknown service": {
			opts:    InsertOptions{Services: []string{"nope"}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			selected, err := selectDocs(docs, tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %t", err, tc.wantErr)
			}

			var got []string
			for _, d := range selected {
				got = append(got, d.path)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("selected %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUpdateDoc(t *testing.T) {
	t.Parallel()

	const section = "## IAM Permissions\n\nnew\n"

	testCases := map[string]struct {
		doc     string // "" for no file
		section string
		want    docResult
	}{
		"inserted":       {doc: "# Title\n", section: section, want: docInserted},
		"updated":        {doc: "# Title\n\n## IAM Permissions\n\nold\n", section: section, want: docUpdated},
		"unchanged":      {doc: "# Title\n\n" + section, section: section, want: docUnchanged},
		"removed":        {doc: "# Title\n\n## IAM Permissions\n\nold\n", want: docRemoved},
		"no permissions": {doc: "# Title", want: docNoPermissions},
		"missing":        {section: section, want: docMissing},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "doc.html.markdown")
			if tc.doc != "" {
				writeFile(t, path, tc.doc)
			}

			got, err := updateDoc(path, tc.section, false)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("result = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestInsertIAMPermissions runs insertion twice on a temporary repository and
// checks the second run changes nothing.
func TestInsertIAMPermissions(t *testing.T) {
	root := t.TempDir()
	resourceDoc := filepath.Join(root, "website/docs/r/test_bucket.html.markdown")
	dataSourceDoc := filepath.Join(root, "website/docs/d/test_bucket.html.markdown")
	writeFile(t, resourceDoc, "# Resource: aws_test_bucket\n\n## Import\n\nText.\n")
	writeFile(t, dataSourceDoc, "# Data Source: aws_test_bucket\n")

	for range 2 {
		if err := InsertIAMPermissions(testPermissionsMapFile, InsertOptions{Root: root, All: true}); err != nil {
			t.Fatal(err)
		}
	}

	wantResource := "# Resource: aws_test_bucket\n\n## Import\n\nText.\n\n" +
		"## IAM Permissions\n\n" +
		"The following permissions are required to manage this resource.\n\n" +
		"### Read (`plan`)\n\n- `s3:PutBucketTagging`\n\n" +
		"### Write (`apply`)\n\n- `s3-object-lambda:PutObject`\n- `s3:GetObject`\n- `s3:PutBucketTagging`\n"
	if got := readFile(t, resourceDoc); got != wantResource {
		t.Errorf("resource doc:\n%s\nwant:\n%s", got, wantResource)
	}

	wantDataSource := "# Data Source: aws_test_bucket\n\n" +
		"## IAM Permissions\n\n" +
		"The following permissions are required to read this data source.\n\n" +
		"### Read (`plan`)\n\n- `s3-object-lambda:PutObject`\n- `s3:GetObject`\n"
	if got := readFile(t, dataSourceDoc); got != wantDataSource {
		t.Errorf("data source doc:\n%s\nwant:\n%s", got, wantDataSource)
	}
}

func TestInsertIAMPermissionsDryRun(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "website/docs/r/test_bucket.html.markdown")
	writeFile(t, path, "# Resource: aws_test_bucket\n")

	opts := InsertOptions{Root: root, Files: []string{"website/docs/r/test_bucket.html.markdown"}, DryRun: true}
	if err := InsertIAMPermissions(testPermissionsMapFile, opts); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, path); got != "# Resource: aws_test_bucket\n" {
		t.Errorf("dry run changed the doc:\n%s", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
