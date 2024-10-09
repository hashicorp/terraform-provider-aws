// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"time"
)

const (
	eventSourceMappingStateCreating  = "Creating"
	eventSourceMappingStateDeleting  = "Deleting"
	eventSourceMappingStateDisabled  = "Disabled"
	eventSourceMappingStateDisabling = "Disabling"
	eventSourceMappingStateEnabled   = "Enabled"
	eventSourceMappingStateEnabling  = "Enabling"
	eventSourceMappingStateUpdating  = "Updating"
)

const (
	iamPropagationTimeout    = 2 * time.Minute
	lambdaPropagationTimeout = 5 * time.Minute // nosemgrep:ci.lambda-in-const-name, ci.lambda-in-var-name
)

type invocationAction string

const (
	invocationActionCreate invocationAction = "create"
	invocationActionDelete invocationAction = "delete"
	invocationActionUpdate invocationAction = "update"
)

type lifecycleScope string

const (
	lifecycleScopeCreateOnly lifecycleScope = "CREATE_ONLY"
	lifecycleScopeCrud       lifecycleScope = "CRUD"
)

func (lifecycleScope) Values() []lifecycleScope {
	return []lifecycleScope{
		lifecycleScopeCreateOnly,
		lifecycleScopeCrud,
	}
}
