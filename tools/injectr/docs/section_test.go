// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docs

import "testing"

func TestRenderSection(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		kind  docKind
		perms docPermissions
		want  string
	}{
		"resource": {
			kind:  resourceKind,
			perms: docPermissions{Read: []string{"s3:GetObject"}, Write: []string{"s3:GetObject", "s3:PutObject"}},
			want: "## IAM Permissions\n\n" +
				"The following permissions are required to manage this resource.\n\n" +
				"### Read (`plan`)\n\n- `s3:GetObject`\n\n" +
				"### Write (`apply`)\n\n- `s3:GetObject`\n- `s3:PutObject`\n",
		},
		"resource without read permissions": {
			kind:  resourceKind,
			perms: docPermissions{Write: []string{"s3:PutObject"}},
			want: "## IAM Permissions\n\n" +
				"The following permissions are required to manage this resource.\n\n" +
				"### Write (`apply`)\n\n- `s3:PutObject`\n",
		},
		"data source": {
			kind:  dataSourceKind,
			perms: docPermissions{Read: []string{"s3:GetObject"}},
			want: "## IAM Permissions\n\n" +
				"The following permissions are required to read this data source.\n\n" +
				"### Read (`plan`)\n\n- `s3:GetObject`\n",
		},
		"no permissions": {kind: resourceKind},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := renderSection(tc.kind.intro, tc.perms); got != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestUpsertSection(t *testing.T) {
	t.Parallel()

	const (
		section = "## IAM Permissions\n\nnew\n"
		old     = "## IAM Permissions\n\nold\n"
	)

	testCases := map[string]struct {
		doc, section, want string
	}{
		"append": {
			doc:     "# Title\n\n## Import\n\nText.\n",
			section: section,
			want:    "# Title\n\n## Import\n\nText.\n\n" + section,
		},
		"append without trailing newline": {
			doc:     "# Title\n\nText.",
			section: section,
			want:    "# Title\n\nText.\n\n" + section,
		},
		"append after trailing blank lines": {
			doc:     "# Title\n\nText.\n\n\n",
			section: section,
			want:    "# Title\n\nText.\n\n" + section,
		},
		"unchanged": {
			doc:     "# Title\n\n" + section,
			section: section,
			want:    "# Title\n\n" + section,
		},
		"update at end": {
			doc:     "# Title\n\n" + old,
			section: section,
			want:    "# Title\n\n" + section,
		},
		"update before another section": {
			doc:     "# Title\n\n" + old + "\n## Import\n\nText.\n",
			section: section,
			want:    "# Title\n\n" + section + "\n## Import\n\nText.\n",
		},
		"remove duplicates": {
			doc:     "# Title\n\n" + old + "\n" + old,
			section: section,
			want:    "# Title\n\n" + section,
		},
		"remove": {
			doc:  "# Title\n\nText.\n\n" + old,
			want: "# Title\n\nText.\n",
		},
		"ignore heading in code block": {
			doc:     "# Title\n\n```markdown\n" + old + "```\n",
			section: section,
			want:    "# Title\n\n```markdown\n" + old + "```\n\n" + section,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := upsertSection(tc.doc, tc.section); got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestHasSection(t *testing.T) {
	t.Parallel()

	if !hasSection("# Title\n\n## IAM Permissions\n") {
		t.Error("section not found")
	}
	if hasSection("# Title\n\n```\n## IAM Permissions\n```\n") {
		t.Error("heading in code block found")
	}
}
