// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docs

import (
	"fmt"
	"iter"
	"strings"
)

const iamHeading = "## IAM Permissions"

// renderSection returns the IAM Permissions section for p, starting with
// intro, or "" if p has no permissions.
func renderSection(intro string, p docPermissions) string {
	if len(p.Read) == 0 && len(p.Write) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n%s\n", iamHeading, intro)
	writeList(&b, "### Read (`plan`)", p.Read)
	writeList(&b, "### Write (`apply`)", p.Write)

	return b.String()
}

func writeList(b *strings.Builder, heading string, perms []string) {
	if len(perms) == 0 {
		return
	}

	fmt.Fprintf(b, "\n%s\n\n", heading)
	for _, perm := range perms {
		fmt.Fprintf(b, "- `%s`\n", perm)
	}
}

// upsertSection returns doc with its IAM Permissions section replaced by
// section, or with section appended if doc has none. An empty section removes
// the existing one. Any duplicate sections are removed.
func upsertSection(doc, section string) string {
	var out strings.Builder
	inSection, replaced := false, false

	for line, isHeading := range h2Lines(doc) {
		if isHeading {
			inSection = strings.TrimSpace(line) == iamHeading
			if inSection && !replaced && section != "" {
				out.WriteString(section + "\n")
				replaced = true
			}
		}
		if !inSection {
			out.WriteString(line)
		}
	}

	result := strings.TrimRight(out.String(), "\n")
	if section != "" && !replaced {
		result += "\n\n" + strings.TrimRight(section, "\n")
	}

	return result + "\n"
}

// hasSection reports whether doc has an IAM Permissions section.
func hasSection(doc string) bool {
	for line, isHeading := range h2Lines(doc) {
		if isHeading && strings.TrimSpace(line) == iamHeading {
			return true
		}
	}
	return false
}

// h2Lines yields each line of doc, with its newline, and whether it is a
// level-2 heading. Lines inside code blocks are never headings.
func h2Lines(doc string) iter.Seq2[string, bool] {
	return func(yield func(string, bool) bool) {
		inCodeBlock := false
		for line := range strings.Lines(doc) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				inCodeBlock = !inCodeBlock
			}
			if !yield(line, !inCodeBlock && strings.HasPrefix(line, "## ")) {
				return
			}
		}
	}
}
