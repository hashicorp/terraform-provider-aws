// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings                   = newBrowserSettingsResource
	ResourceBrowserSettingsAssociation        = newBrowserSettingsAssociationResource
	ResourceDataProtectionSettings            = newDataProtectionSettingsResource
	ResourceDataProtectionSettingsAssociation = newDataProtectionSettingsAssociationResource
	ResourceIdentityProvider                  = newIdentityProviderResource
	ResourceIPAccessSettings                  = newIPAccessSettingsResource
	ResourceIPAccessSettingsAssociation       = newIPAccessSettingsAssociationResource
	ResourceNetworkSettings                   = newNetworkSettingsResource
	ResourcePortal                            = newPortalResource
	ResourceTrustStore                        = newTrustStoreResource
	ResourceUserAccessLoggingSettings         = newUserAccessLoggingSettingsResource
	ResourceUserSettings                      = newUserSettingsResource

	FindBrowserSettingsByARN           = findBrowserSettingsByARN
	FindDataProtectionSettingsByARN    = findDataProtectionSettingsByARN
	FindIdentityProviderByARN          = findIdentityProviderByARN
	FindIPAccessSettingsByARN          = findIPAccessSettingsByARN
	FindNetworkSettingsByARN           = findNetworkSettingsByARN
	FindPortalByARN                    = findPortalByARN
	FindTrustStoreByARN                = findTrustStoreByARN
	FindUserAccessLoggingSettingsByARN = findUserAccessLoggingSettingsByARN
	FindUserSettingsByARN              = findUserSettingsByARN

	PortalARNFromIdentityProviderARN = portalARNFromIdentityProviderARN
)
