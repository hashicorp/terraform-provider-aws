// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

const (
	// If AWS adds these to the API, we should use those and remove these.

	acceptLanguageEnglish  = "en"
	acceptLanguageJapanese = "jp"
	acceptLanguageChinese  = "zh"

	constraintTypeLaunch         = "LAUNCH"
	constraintTypeNotification   = "NOTIFICATION"
	constraintTypeResourceUpdate = "RESOURCE_UPDATE"
	constraintTypeStackset       = "STACKSET"
	constraintTypeTemplate       = "TEMPLATE"
)

func acceptLanguage_Values() []string {
	return []string{
		acceptLanguageEnglish,
		acceptLanguageJapanese,
		acceptLanguageChinese,
	}
}

func constraintType_Values() []string {
	return []string{
		constraintTypeLaunch,
		constraintTypeNotification,
		constraintTypeResourceUpdate,
		constraintTypeStackset,
		constraintTypeTemplate,
	}
}
