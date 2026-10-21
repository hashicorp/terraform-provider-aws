// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Seed a known roster so classification is deterministic in tests.
	maintainers = map[string]bool{"gdavison": true, "jar-b": true, "breathingdust": true}
	os.Exit(m.Run())
}

func TestOwner(t *testing.T) {
	tests := []struct {
		name, prAuthor, cemail, wantStatus string
	}{
		{"maintainer via PR author", "gdavison", "someone@example.com", "maintainer"},
		{"maintainer via email", "", "jared.baker@hashicorp.com", "maintainer"},
		{"maintainer via email (mgr)", "", "sdavis@hashicorp.com", "maintainer"},
		{"former employee", "", "sharon.nam@hashicorp.com", "other"},
		{"bot app author", "app/oss-core-libraries-dashboard", "", "bot"},
		{"bot committer", "", "41898282+github-actions[bot]@users.noreply.github.com", "bot"},
		{"external community", "mattbork", "", "other"},
		{"unknown", "", "nobody@nowhere.local", "unknown"},
		{"prefers maintainer over bot committer", "gdavison", "41898282+github-actions[bot]@users.noreply.github.com", "maintainer"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := owner(tc.prAuthor, tc.cemail); got != tc.wantStatus {
				t.Errorf("owner(%q, %q) = %q, want %q", tc.prAuthor, tc.cemail, got, tc.wantStatus)
			}
		})
	}
}

func TestBucketOf(t *testing.T) {
	tests := []struct {
		name string
		b    branch
		want string
	}{
		{"gh-pages protected", branch{name: "gh-pages"}, "X2"},
		{"release protected", branch{name: "release/6.0.0-beta", inMain: true}, "X1"},
		{"v-version protected", branch{name: "v6.0.0-upgrade", inMain: true}, "X1"},
		{"open PR is active", branch{name: "f-thing", prState: "OPEN", ownerStatus: "maintainer"}, "R6"},
		{"bot closed", branch{name: "compliance/x", prState: "CLOSED", ownerStatus: "bot"}, "R1"},
		{"in-main lossless", branch{name: "f-merged", inMain: true, ownerStatus: "maintainer"}, "R2"},
		{"no current owner", branch{name: "f-old", ownerStatus: "other", age: 100}, "R3"},
		{"maintainer stale", branch{name: "f-stale", ownerStatus: "maintainer", age: 800}, "R4"},
		{"maintainer recent", branch{name: "f-recent", ownerStatus: "maintainer", age: 30}, "R5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := bucketOf(tc.b); got != tc.want {
				t.Errorf("bucketOf(%+v) = %q, want %q", tc.b, got, tc.want)
			}
		})
	}
}

func TestLoadMaintainers(t *testing.T) {
	if len(loadMaintainers("does-not-exist")) != 0 {
		t.Error("expected empty set for missing file")
	}
}
