// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	gversion "github.com/YakDriver/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// compareActualEngineVersion sets engine version related attributes
//
// `engine_version_actual` is always set to newVersion
//
// `engine_version` is set to newVersion unless:
//   - old and pending versions are equal (ie. the update is awaiting a
//     maintenance window)
//   - old and new versions are not exactly equal, but match after accounting
//     for an omitted patch value in the configuration (ie. old="1.3",
//     new="1.3.27" will not trigger a set)
func compareActualEngineVersion(d *schema.ResourceData, oldVersion, newVersion, pendingVersion string) {
	d.Set("engine_version_actual", newVersion)

	if oldVersion != "" && oldVersion == pendingVersion {
		return
	}

	newVersionSubstr := newVersion
	if len(newVersion) > len(oldVersion) {
		newVersionSubstr = string([]byte(newVersion)[0 : len(oldVersion)+1])
	}
	if oldVersion != newVersion && oldVersion+"." == newVersionSubstr {
		return
	}

	d.Set(names.AttrEngineVersion, newVersion)
}

// engineVersionIsNewer returns true when v1 is strictly greater than v2 using
// numeric-aware version comparison. This handles both plain versions ("16.3")
// and Aurora-style versions ("8.0.mysql_aurora.3.10.0") correctly, including
// unpadded minor/patch segments where lexicographic comparison would fail
// (e.g. "3.9" < "3.10" even though "3.9" > "3.10" as a string).
// Returns false if either string cannot be parsed as a version.
func engineVersionIsNewer(v1, v2 string) bool {
	return !gversion.LessThan(v1, v2) && v1 != v2
}
