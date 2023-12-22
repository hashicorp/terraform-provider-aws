// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

// Exports for use in tests only.
var (
	ResourceDataLake  = newDataLakeResource
	ResourceLogSource = newResourceLogSource

	FindDataLakeByARN             = findDataLakeByARN
	FindLogSourceById             = findLogSourceById
	ExtractLogSourceConfiguration = extractLogSourceConfiguration
)
