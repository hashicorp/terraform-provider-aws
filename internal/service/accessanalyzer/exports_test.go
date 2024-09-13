// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer

// Exports for use in tests only.
var (
	ArchiveRuleParseResourceID  = archiveRuleParseResourceID
	FindAnalyzerByName          = findAnalyzerByName
	FindArchiveRuleByTwoPartKey = findArchiveRuleByTwoPartKey

	ResourceAnalyzer    = resourceAnalyzer
	ResourceArchiveRule = resourceArchiveRule
)
