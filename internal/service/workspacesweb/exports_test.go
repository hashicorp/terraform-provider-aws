// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

// Exports for use in tests only.
var (
	ResourceBrowserSettings                      = newBrowserSettingsResource
	ResourceBrowserSettingsAssociation           = newBrowserSettingsAssociationResource
	ResourceDataProtectionSettings               = newDataProtectionSettingsResource
	ResourceDataProtectionSettingsAssociation    = newDataProtectionSettingsAssociationResource
	ResourceIdentityProvider                     = newIdentityProviderResource
	ResourceIPAccessSettings                     = newIPAccessSettingsResource
	ResourceIPAccessSettingsAssociation          = newIPAccessSettingsAssociationResource
	ResourceNetworkSettings                      = newNetworkSettingsResource
	ResourceNetworkSettingsAssociation           = newNetworkSettingsAssociationResource
	ResourcePortal                               = newPortalResource
	ResourceSessionLogger                        = newSessionLoggerResource
	ResourceSessionLoggerAssociation             = newSessionLoggerAssociationResource
	ResourceTrustStore                           = newTrustStoreResource
	ResourceTrustStoreAssociation                = newTrustStoreAssociationResource
	ResourceUserAccessLoggingSettings            = newUserAccessLoggingSettingsResource
	ResourceUserAccessLoggingSettingsAssociation = newUserAccessLoggingSettingsAssociationResource
	ResourceUserSettings                         = newUserSettingsResource
	ResourceUserSettingsAssociation              = newUserSettingsAssociationResource

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

	PortalARNFromIdentityProviderARN = portalARNFromIdentityProviderARN
)
