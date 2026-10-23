// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

// Exports for use in tests only.
var (
	ResourceApplication         = resourceApplication
	ResourceApplicationSnapshot = resourceApplicationSnapshot
	ResourceSnapshot            = newSnapshotResource
	ResourceSnapshotImport      = newSnapshotImportResource

	FindApplicationDetailByName     = findApplicationDetailByName
	FindEnhancedSnapshotByName      = findEnhancedSnapshotByName
	FindSnapshotDetailsByTwoPartKey = findSnapshotDetailsByTwoPartKey

	RuntimeSupportsEnhancedSnapshots = runtimeSupportsEnhancedSnapshots
)
