// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

// Exports for use in tests only.
var (
	ResourceActivity     = resourceActivity
	ResourceAlias        = resourceAlias
	ResourceStateMachine = resourceStateMachine

	FindActivityByARN     = findActivityByARN
	FindAliasByARN        = findAliasByARN
	FindStateMachineByARN = findStateMachineByARN
)
