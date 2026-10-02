// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

// Exports for use in tests only.
var (
	ResourceRule     = newRuleResource
	ResourceTemplate = newTemplateResource

	FindRuleByARN     = findRuleByARN
	FindTemplateByARN = findTemplateByARN
)
