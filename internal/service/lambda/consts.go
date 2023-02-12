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
	lambdaInvocationActionCreate = "create"
	lambdaInvocationActionDelete = "delete"
	lambdaInvocationActionUpdate = "update"
)

const (
	lambdaLifecycleScopeCreateOnly = "CREATE_ONLY"
	lambdaLifecycleScopeCrud       = "CRUD"
)

func lambdaLifecycleScope_Values() []string {
	return []string{
		lambdaLifecycleScopeCreateOnly,
		lambdaLifecycleScopeCrud,
	}
}
