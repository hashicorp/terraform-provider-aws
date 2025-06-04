// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings           = newBrowserSettingsResource
	ResourceNetworkSettings           = newNetworkSettingsResource
	ResourceUserSettings              = newUserSettingsResource
	ResourceDataProtectionSettings    = newDataProtectionSettingsResource
	ResourceIPAccessSettings          = newIPAccessSettingsResource
	ResourceUserAccessLoggingSettings = newUserAccessLoggingSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
)
