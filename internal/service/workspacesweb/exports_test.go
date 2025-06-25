// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings           = newBrowserSettingsResource
	ResourceDataProtectionSettings    = newDataProtectionSettingsResource
	ResourceIPAccessSettings          = newIPAccessSettingsResource
	ResourceNetworkSettings           = newNetworkSettingsResource
	ResourceUserAccessLoggingSettings = newUserAccessLoggingSettingsResource
	ResourceUserSettings              = newUserSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN
)
