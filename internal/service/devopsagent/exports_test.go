// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

// Exports for use in tests only.
var (
	ResourceAgentSpace = newAgentSpaceResource
	ResourceAsset      = newAssetResource
	ResourceAssetFile  = newAssetFileResource

	FindAgentSpaceByID  = findAgentSpaceByID
	FindAssetByID       = findAssetByID
	FindAssetFileByPath = findAssetFileByPath
)
