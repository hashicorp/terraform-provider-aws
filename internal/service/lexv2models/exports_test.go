// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

// Exports for use in tests only.
var (
	ResourceBot        = newBotResource
	ResourceBotLocale  = newBotLocaleResource
	ResourceBotVersion = newBotVersionResource
	ResourceIntent     = newIntentResource
	ResourceSlot       = newSlotResource
	ResourceSlotType   = newSlotTypeResource

	FindBotByID                 = findBotByID
	FindBotLocaleByThreePartKey = findBotLocaleByThreePartKey
	FindBotVersionByTwoPartKey  = findBotVersionByTwoPartKey
	FindSlotByID                = findSlotByID

	IntentFlexOpt = intentFlexOpt
)
