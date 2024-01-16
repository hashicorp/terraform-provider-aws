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
	propagationTimeout = 5 * time.Minute
)

const (
	invocationActionCreate = "create"
	invocationActionDelete = "delete"
	invocationActionUpdate = "update"
)

const (
	lifecycleScopeCreateOnly = "CREATE_ONLY"
	lifecycleScopeCrud       = "CRUD"
)

func lifecycleScope_Values() []string {
	return []string{
		lifecycleScopeCreateOnly,
		lifecycleScopeCrud,
	}
}
