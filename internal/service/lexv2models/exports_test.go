// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

// Exports for use in tests only.
var (
	ResourceBot        = newResourceBot
	ResourceBotLocale  = newResourceBotLocale
	ResourceBotVersion = newResourceBotVersion
	ResourceIntent     = newResourceIntent
	ResourceSlot       = newResourceSlot
	ResourceSlotType   = newResourceSlotType

	FindSlotByID = findSlotByID
)
