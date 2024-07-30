// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfprotov5

// ServerCapabilities allows providers to communicate optionally supported
// protocol features, such as forward-compatible Terraform behavior changes.
//
// This information is used in GetProviderSchemaResponse as capabilities are
// static features which must be known upfront in the provider server.
type ServerCapabilities struct {
	// GetProviderSchemaOptional signals that this provider does not require
	// having the GetProviderSchema RPC called first to operate normally. This
	// means the caller can use a cached copy of the provider's schema instead.
	GetProviderSchemaOptional bool

	// MoveResourceState signals that a provider supports the MoveResourceState
	// RPC.
	MoveResourceState bool

	// PlanDestroy signals that a provider expects a call to
	// PlanResourceChange when a resource is going to be destroyed. This is
	// opt-in to prevent unexpected errors or panics since the
	// ProposedNewState in PlanResourceChangeRequest will be a null value.
	PlanDestroy bool
}
