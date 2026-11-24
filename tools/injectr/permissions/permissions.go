// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package permissions maps the AWS SDK calls in sdkmap's output to the IAM
// permissions they require, using the AWS Service Reference.
package permissions

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/internal/atomicfile"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/serviceref"
)

// PermissionStats summarizes an AddPermissions run.
type PermissionStats struct {
	FuncRefs int // function references processed
	SDKCalls int // SDK calls processed, counting each occurrence
	Resolved int // SDK calls mapped to IAM permissions

	Unresolved map[unresolvedReason]int
	// UnresolvedByPackage counts unresolved SDK calls by AWS SDK for Go v2
	// package, e.g. "qldb".
	UnresolvedByPackage map[string]int
}

// GeneratePermissionsMap reads the service reference data in serviceDataDir
// and the sdkmap output at sdkMapPath, and writes the sdkmap output with IAM
// permissions added to outPath.
func GeneratePermissionsMap(serviceDataDir, sdkMapPath, outPath string) error {
	start := time.Now()
	refs, err := serviceref.LoadServiceRefs(serviceDataDir)
	if err != nil {
		return err
	}
	index := NewOperationIndex(refs)
	fmt.Printf("#### Indexed %d service references from %s (%s)\n", len(refs), serviceDataDir, since(start))

	start = time.Now()
	m, err := LoadSDKMap(sdkMapPath)
	if err != nil {
		return err
	}
	fmt.Printf("#### Loaded %d resources and %d data sources from %s (%s)\n", len(m.Resources), len(m.DataSources), sdkMapPath, since(start))

	start = time.Now()
	stats := AddPermissions(m, index)
	fmt.Printf("#### Added IAM permissions (%s)\n", since(start))

	start = time.Now()
	if err := WritePermissionsMap(outPath, m); err != nil {
		return err
	}
	fmt.Printf("#### Wrote %s (%s)\n", outPath, since(start))

	printPermissionStats(stats)

	return nil
}

// AddPermissions sets the IAM permissions, and any unresolved SDK calls, on
// every function reference in m.
func AddPermissions(m *PermissionsMap, index *OperationIndex) PermissionStats {
	stats := PermissionStats{
		Unresolved:          make(map[unresolvedReason]int),
		UnresolvedByPackage: make(map[string]int),
	}

	for f := range m.FuncRefs() {
		stats.FuncRefs++
		addPermissions(f, index, &stats)
	}

	return stats
}

func addPermissions(f *FuncRef, index *OperationIndex, stats *PermissionStats) {
	perms := []string{} // non-nil, so that none is written as []
	f.UnresolvedSDKCalls = nil

	for _, call := range f.SDKCalls {
		stats.SDKCalls++

		callPerms, reason := index.Resolve(call)
		if reason != "" {
			f.UnresolvedSDKCalls = append(f.UnresolvedSDKCalls, UnresolvedSDKCall{SDKCall: call, Reason: reason})
			stats.Unresolved[reason]++
			goSDKPackage, _, _ := strings.Cut(call, ".")
			stats.UnresolvedByPackage[goSDKPackage]++
			continue
		}

		stats.Resolved++
		perms = append(perms, callPerms...)
	}

	slices.Sort(perms)
	f.IAMPermissions = slices.Compact(perms)
}

// WritePermissionsMap writes m to path, atomically, so that an interrupted run
// never leaves a partial file at path.
func WritePermissionsMap(path string, m *PermissionsMap) error {
	b, err := m.encode()
	if err != nil {
		return err
	}

	return atomicfile.Write(path, b)
}

func printPermissionStats(stats PermissionStats) {
	unresolved := stats.SDKCalls - stats.Resolved
	fmt.Printf("#### Function references: %d, SDK calls: %d, Resolved: %d (%s), Unresolved: %d\n",
		stats.FuncRefs, stats.SDKCalls, stats.Resolved, percent(stats.Resolved, stats.SDKCalls), unresolved)

	for _, reason := range slices.Sorted(maps.Keys(stats.Unresolved)) {
		fmt.Printf("  %-27s %d\n", reason, stats.Unresolved[reason])
	}

	const top = 10
	packages := slices.SortedFunc(maps.Keys(stats.UnresolvedByPackage), func(a, b string) int {
		return cmp.Or(cmp.Compare(stats.UnresolvedByPackage[b], stats.UnresolvedByPackage[a]), cmp.Compare(a, b))
	})
	if len(packages) > 0 {
		fmt.Printf("#### Unresolved SDK calls by package (top %d of %d)\n", min(top, len(packages)), len(packages))
	}
	for _, pkg := range packages[:min(top, len(packages))] {
		fmt.Printf("  %-27s %d\n", pkg, stats.UnresolvedByPackage[pkg])
	}
}

func percent(n, total int) string {
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(n)/float64(total))
}

func since(start time.Time) string {
	return time.Since(start).Round(time.Millisecond).String()
}
