// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexv2models

// Exports for use in tests only.
var (
	ResourceBot        = newBotResource
	ResourceBotAlias   = newBotAliasResource
	ResourceBotLocale  = newBotLocaleResource
	ResourceBotVersion = newBotVersionResource
	ResourceIntent     = newIntentResource
	ResourceSlot       = newSlotResource
	ResourceSlotType   = newSlotTypeResource

	FindBotAliasByTwoPartKey    = findBotAliasByTwoPartKey
	FindBotByID                 = findBotByID
	FindBotLocaleByThreePartKey = findBotLocaleByThreePartKey
	FindBotVersionByTwoPartKey  = findBotVersionByTwoPartKey
	FindSlotByID                = findSlotByID

	IntentFlexOpt = intentFlexOpt

	ArePromptAttemptsEqual             = arePromptAttemptsEqual
	DefaultPromptAttemptsSpecification = defaultPromptAttemptsSpecification
)
