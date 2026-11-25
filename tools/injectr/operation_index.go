// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"slices"
	"strings"
)

// unresolvedReason explains why an SDK call could not be mapped to IAM
// permissions.
type unresolvedReason string

const (
	// No service reference lists any operation for the call's SDK client:
	// either the service is missing, or its data has no Operations.
	reasonServiceHasNoOperations unresolvedReason = "service-has-no-operations"
	// The service lists operations, but not this one.
	reasonOperationNotFound unresolvedReason = "operation-not-found"
	// The operation is listed, but requires no IAM actions.
	reasonNoIAMActions unresolvedReason = "no-iam-actions"
)

// operationKey identifies an operation by its normalized Boto3 client name and
// operation name, e.g. {"elbv2", "CreateLoadBalancer"}.
type operationKey struct {
	client    string
	operation string
}

// OperationIndex looks up the IAM permissions required by an SDK call in
// constant time.
type OperationIndex struct {
	// permissions holds each operation's IAM actions, sorted and deduplicated,
	// in policy form, e.g. "s3:GetObject". An operation that requires none
	// maps to an empty slice.
	permissions map[operationKey][]string
	// clients holds every client with at least one operation.
	clients map[string]struct{}
}

// NewOperationIndex indexes the operations of every service in refs by the
// Boto3 client names that implement them.
func NewOperationIndex(refs ServiceRefs) *OperationIndex {
	idx := &OperationIndex{
		permissions: make(map[operationKey][]string),
		clients:     make(map[string]struct{}),
	}

	for _, ref := range refs {
		for _, op := range ref.Operations {
			for _, sdk := range op.SDK {
				// Service reference data only lists Boto3 clients today. Skip
				// any other SDK, whose client names may not follow Boto3's.
				if sdk.Package != "Boto3" {
					continue
				}
				key := operationKey{client: normalizeClient(sdk.Name), operation: op.Name}
				idx.clients[key.client] = struct{}{}
				idx.permissions[key] = append(idx.permissions[key], iamActions(op)...)
			}
		}
	}

	// Services are visited in map order, so sort to make the result stable if
	// more than one lists the same operation.
	for key, perms := range idx.permissions {
		slices.Sort(perms)
		idx.permissions[key] = slices.Compact(perms)
	}

	return idx
}

// Resolve returns the IAM permissions required by sdkCall, an sdkmap call such
// as "elasticloadbalancingv2.CreateLoadBalancer". If none can be found, it
// returns the reason instead. The returned slice must not be modified.
func (idx *OperationIndex) Resolve(sdkCall string) ([]string, unresolvedReason) {
	goSDKPackage, operation, _ := strings.Cut(sdkCall, ".")
	key := operationKey{client: boto3Client(goSDKPackage), operation: operation}

	perms, ok := idx.permissions[key]
	switch {
	case ok && len(perms) > 0:
		return perms, ""
	case ok:
		return nil, reasonNoIAMActions
	}

	if _, ok := idx.clients[key.client]; ok {
		return nil, reasonOperationNotFound
	}
	return nil, reasonServiceHasNoOperations
}

func iamActions(op Operation) []string {
	actions := make([]string, 0, len(op.AuthorizedActions))
	for _, aa := range op.AuthorizedActions {
		actions = append(actions, aa.IAMAction())
	}
	return actions
}
