// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

const (
	configurationProfileTypeFeatureFlags = "AWS.AppConfig.FeatureFlags"
	configurationProfileTypeFreeform     = "AWS.Freeform"
)

func ConfigurationProfileType_Values() []string {
	return []string{
		configurationProfileTypeFeatureFlags,
		configurationProfileTypeFreeform,
	}
}
