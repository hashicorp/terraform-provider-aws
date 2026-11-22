// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// InsertOptions selects the documentation to add IAM permissions to.
type InsertOptions struct {
	Root     string   // repository root
	Files    []string // documentation files, relative to Root
	Services []string // service packages, e.g. "s3"
	All      bool
	DryRun   bool // report changes without writing them
}

func (o InsertOptions) Enabled() bool {
	return o.All || len(o.Files) > 0 || len(o.Services) > 0
}

// doc is a documentation file and the IAM Permissions section it should
// contain.
type doc struct {
	path    string // relative to the repository root, e.g. "website/docs/r/s3_bucket.html.markdown"
	service string
	section string // empty if there are no permissions to list
}

type docResult string

const (
	docInserted      docResult = "inserted"
	docUpdated       docResult = "updated"
	docRemoved       docResult = "removed"
	docUnchanged     docResult = "unchanged"
	docNoPermissions docResult = "no permissions"
	docMissing       docResult = "no doc file"
)

// InsertIAMPermissions adds an IAM Permissions section, built from the
// permissions map at permissionsMapPath, to the documentation selected by opts.
func InsertIAMPermissions(permissionsMapPath string, opts InsertOptions) error {
	m, err := LoadSDKMap(permissionsMapPath)
	if err != nil {
		return err
	}

	docs, err := selectDocs(allDocs(m), opts)
	if err != nil {
		return err
	}

	counts := make(map[docResult]int)
	for _, d := range docs {
		result, err := updateDoc(filepath.Join(opts.Root, d.path), d.section, opts.DryRun)
		if err != nil {
			return err
		}
		counts[result]++
		if result == docInserted || result == docUpdated || result == docRemoved {
			fmt.Printf("%s %s\n", result, d.path)
		}
	}

	var summary []string
	for _, result := range slices.Sorted(maps.Keys(counts)) {
		summary = append(summary, fmt.Sprintf("%s: %d", result, counts[result]))
	}
	dryRun := ""
	if opts.DryRun {
		dryRun = " (dry run)"
	}
	fmt.Printf("#### Docs%s: %s\n", dryRun, strings.Join(summary, ", "))

	return nil
}

// allDocs returns the documentation of every type in map.
func allDocs(m *PermissionsMap) []doc {
	docs := make([]doc, 0, len(m.Resources)+len(m.DataSources))
	for i := range m.Resources {
		r := &m.Resources[i]
		docs = append(docs, resourceKind.newDoc(r.ResourceType, r.Service, resourcePermissions(r)))
	}
	for i := range m.DataSources {
		d := &m.DataSources[i]
		docs = append(docs, dataSourceKind.newDoc(d.DataSourceType, d.Service, dataSourcePermissions(d)))
	}
	return docs
}

// selectDocs returns the docs selected by opts, each at most once.
func selectDocs(docs []doc, opts InsertOptions) ([]doc, error) {
	if opts.All {
		return docs, nil
	}

	var selected []doc
	seen := make(map[string]bool)
	add := func(d doc) {
		if !seen[d.path] {
			seen[d.path] = true
			selected = append(selected, d)
		}
	}

	for _, file := range opts.Files {
		i := slices.IndexFunc(docs, func(d doc) bool { return d.path == filepath.ToSlash(filepath.Clean(file)) })
		if i < 0 {
			return nil, fmt.Errorf("%s: not the documentation of any resource or data source", file)
		}
		add(docs[i])
	}

	for _, service := range opts.Services {
		found := false
		for _, d := range docs {
			if d.service == service {
				add(d)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("%s: no resources or data sources in this service", service)
		}
	}

	return selected, nil
}

// updateDoc sets the IAM Permissions section of the file at path to section,
// writing the file only if it changes.
func updateDoc(path, section string, dryRun bool) (docResult, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return docMissing, nil
	}
	if err != nil {
		return "", err
	}

	old := string(b)
	hadSection := hasSection(old)
	if section == "" && !hadSection {
		return docNoPermissions, nil
	}

	updated := upsertSection(old, section)
	if updated == old {
		return docUnchanged, nil
	}

	if !dryRun {
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return "", err
		}
	}

	switch {
	case section == "":
		return docRemoved, nil
	case hadSection:
		return docUpdated, nil
	default:
		return docInserted, nil
	}
}
