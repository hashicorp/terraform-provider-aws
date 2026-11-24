// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import "slices"

// docPermissions are the IAM permissions listed in a type's documentation.
type docPermissions struct {
	Read  []string // needed to plan
	Write []string // needed to apply; includes Read. Nil for kinds that are only read, e.g. data sources.
}

func resourcePermissions(r *Resource) docPermissions {
	var tagsList, tagsUpdate *FuncRef
	if r.Tags != nil {
		tagsList, tagsUpdate = r.Tags.List, r.Tags.Update
	}

	read := []*FuncRef{r.Read, r.CustomizeDiff, r.ModifyPlan, r.Import, tagsList}
	write := append(slices.Clone(read), r.Create, r.Update, r.Delete, tagsUpdate)

	return docPermissions{Read: permissionsOf(read), Write: permissionsOf(write)}
}

func dataSourcePermissions(d *DataSource) docPermissions {
	refs := []*FuncRef{d.Read}
	if d.Tags != nil {
		refs = append(refs, d.Tags.List)
	}

	return docPermissions{Read: permissionsOf(refs)}
}

// permissionsOf returns the sorted, deduplicated IAM permissions of refs,
// including their nested functions. Nil refs are skipped.
func permissionsOf(refs []*FuncRef) []string {
	var perms []string
	var walk func([]*FuncRef)
	walk = func(refs []*FuncRef) {
		for _, f := range refs {
			if f == nil {
				continue
			}
			perms = append(perms, f.IAMPermissions...)
			walk(f.Functions)
		}
	}
	walk(refs)

	slices.Sort(perms)
	return slices.Compact(perms)
}
