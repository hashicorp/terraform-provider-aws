// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

// Exports for use in other modules.
var (
	FindStackByName      = findStackByName
	FindTypeByName       = findTypeByName
	WaitChangeSetCreated = waitChangeSetCreated
	WaitStackCreated     = waitStackCreated
	WaitStackDeleted     = waitStackDeleted
	WaitStackUpdated     = waitStackUpdated
)
