// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package chatbot

// Exports for use in tests only.
var (
	ResourceSlackChannelConfiguration = newSlackChannelConfigurationResource
	ResourceTeamsChannelConfiguration = newTeamsChannelConfigurationResource

	FindSlackChannelConfigurationByARN    = findSlackChannelConfigurationByARN
	FindTeamsChannelConfigurationByTeamID = findTeamsChannelConfigurationByTeamID
)
