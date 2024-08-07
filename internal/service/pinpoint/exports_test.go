// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

// Exports for use in tests only.
var (
	ResourceApp          = resourceApp
	ResourceEmailChannel = resourceEmailChannel
	ResourceEventStream  = resourceEventStream
	ResourceSMSChannel   = resourceSMSChannel

	FindADMChannelByApplicationId             = findADMChannelByApplicationId
	FindAPNSChannelByApplicationId            = findAPNSChannelByApplicationId
	FindAPNSSandboxChannelByApplicationId     = findAPNSSandboxChannelByApplicationId
	FindAPNSVoIPChannelByApplicationId        = findAPNSVoIPChannelByApplicationId
	FindAPNSVoIPSandboxChannelByApplicationId = findAPNSVoIPSandboxChannelByApplicationId
	FindAppByID                               = findAppByID
	FindBaiduChannelByApplicationId           = findBaiduChannelByApplicationId
	FindEmailChannelByApplicationId           = findEmailChannelByApplicationId
	FindEventStreamByApplicationId            = findEventStreamByApplicationId
	FindGCMChannelByApplicationId             = findGCMChannelByApplicationId
	FindSMSChannelByApplicationId             = findSMSChannelByApplicationId
)
