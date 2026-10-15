// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"regexp"
)

// CIDoc is the parsed view of docs/continuous-integration.md.
//
// We only care about which make targets are referenced from the document,
// so the parsed view is just a set of names. A target is "referenced" if
// the document contains `make <target>` anywhere -- in code blocks, prose,
// or section headings.
type CIDoc struct {
	Refs map[string]bool
}

// reMakeRef matches `make <target>` references. The negative-lookbehind for
// `e` is unsupported in RE2, so we constrain via boundary characters and
// rely on Go's `regexp` keeping the match minimal.
//
// Examples it captures:
//
//	make ci
//	`make ci-quick`
//	  $ make gh-workflow-lint
var reMakeRef = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_-])make[ \t]+([a-zA-Z][a-zA-Z0-9_-]*)\b`)

// ParseCIDoc reads the continuous-integration document and collects every
// make target referenced anywhere within it.
func ParseCIDoc(path string) (*CIDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := &CIDoc{Refs: make(map[string]bool)}
	for _, m := range reMakeRef.FindAllSubmatch(b, -1) {
		doc.Refs[string(m[1])] = true
	}
	return doc, nil
}
