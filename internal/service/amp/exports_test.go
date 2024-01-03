// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

// Exports for use in tests only.
var (
	ResourceAlertManagerDefinition = resourceAlertManagerDefinition
	ResourceScraper                = newResourceScraper

	FindAlertManagerDefinitionByID = findAlertManagerDefinitionByID
)
