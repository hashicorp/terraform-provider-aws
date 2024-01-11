// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

// Exports for use in tests only.
var (
	ResourceAWSLogSource = newAWSLogSourceResource
	ResourceDataLake     = newDataLakeResource

	FindAWSLogSourceBySourceName = findAWSLogSourceBySourceName
	FindDataLakeByARN            = findDataLakeByARN
)
