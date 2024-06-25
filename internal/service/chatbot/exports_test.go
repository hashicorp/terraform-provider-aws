// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

// Exports for use in tests only.
var (
	ResourceSlackChannelConfiguration = newSlackChannelConfigurationResource

	FindSlackChannelConfigurationByID = findSlackChannelConfigurationByID
)
