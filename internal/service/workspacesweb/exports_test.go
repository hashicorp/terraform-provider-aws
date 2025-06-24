// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings           = newBrowserSettingsResource
	ResourceDataProtectionSettings    = newDataProtectionSettingsResource
	ResourceIPAccessSettings          = newIPAccessSettingsResource
	ResourceNetworkSettings           = newNetworkSettingsResource
<<<<<<< HEAD
=======
	ResourceTrustStore                = newTrustStoreResource
>>>>>>> 414521b8b0 (Added resource aws_workspacesweb_trust_store.)
	ResourceUserAccessLoggingSettings = newUserAccessLoggingSettingsResource
	ResourceUserSettings              = newUserSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
<<<<<<< HEAD
=======
	FindTrustStoreByARN                = findTrustStoreByARN
>>>>>>> 414521b8b0 (Added resource aws_workspacesweb_trust_store.)
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN
)
