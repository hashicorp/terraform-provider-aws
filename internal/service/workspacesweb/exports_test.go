// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings = newBrowserSettingsResource
	ResourceNetworkSettings = newNetworkSettingsResource

	FindBrowserSettingsByARN = findBrowserSettingsByARN
	FindNetworkSettingsByARN = findNetworkSettingsByARN
)
