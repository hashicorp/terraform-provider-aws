// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings           = newBrowserSettingsResource
	ResourceDataProtectionSettings    = newDataProtectionSettingsResource
	ResourceIdentityProvider          = newIdentityProviderResource
	ResourceIPAccessSettings          = newIPAccessSettingsResource
	ResourceNetworkSettings           = newNetworkSettingsResource
	ResourcePortal                    = newPortalResource
	ResourceSessionLogger             = newSessionLoggerResource
	ResourceTrustStore                = newTrustStoreResource
	ResourceUserAccessLoggingSettings = newUserAccessLoggingSettingsResource
	ResourceUserSettings              = newUserSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIdentityProviderByARN          = findIdentityProviderByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
	FindPortalByARN                    = findPortalByARN
	FindSessionLoggerByARN             = findSessionLoggerByARN
	FindTrustStoreByARN                = findTrustStoreByARN
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN
)
