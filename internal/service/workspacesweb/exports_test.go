// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings           = newBrowserSettingsResource
	ResourceDataProtectionSettings    = newDataProtectionSettingsResource
	ResourceIPAccessSettings          = newIPAccessSettingsResource
	ResourceNetworkSettings           = newNetworkSettingsResource
	ResourcePortal                    = newPortalResource
	ResourceTrustStore                = newTrustStoreResource
	ResourceUserAccessLoggingSettings = newUserAccessLoggingSettingsResource
	ResourceUserSettings              = newUserSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
	FindPortalByARN                    = findPortalByARN
	FindTrustStoreByARN                = findTrustStoreByARN
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN
)
