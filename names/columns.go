// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

const (
	ColAWSCLIV2Command         = 0
	ColAWSCLIV2CommandNoDashes = 1
	ColGoV1Package             = 2
	ColGoV2Package             = 3
	ColProviderPackageActual   = 4
	ColProviderPackageCorrect  = 5
	ColSplitPackageRealPackage = 6
	ColAliases                 = 7
	ColProviderNameUpper       = 8
	ColGoV1ClientTypeName      = 9
	ColSkipClientGenerate      = 10
	ColClientSDKV1             = 11
	ColClientSDKV2             = 12
	ColResourcePrefixActual    = 13
	ColResourcePrefixCorrect   = 14
	ColFilePrefix              = 15
	ColDocPrefix               = 16
	ColHumanFriendly           = 17
	ColBrand                   = 18
	ColExclude                 = 19 // If set, the service is completely ignored
	ColNotImplemented          = 20 // If set, the service will be included in, e.g. labels, but not have a service client
	ColAllowedSubcategory      = 21
	ColDeprecatedEnvVar        = 22
	ColEnvVar                  = 23
	ColNote                    = 24
)
