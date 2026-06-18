// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"github.com/hashicorp/go-cty/cty"
)

// ResourceDiffer exposes the interface for accessing changes in a resource
// Implementations:
// * schema.ResourceData
// * schema.ResourceDiff
// Matches the public part of helper/schema/resourceDiffer:
// https://github.com/hashicorp/terraform-plugin-sdk/blob/28e631776d97f0a5a5942b3524814addbef90875/helper/schema/schema.go#L1104-L1112
type ResourceDiffer interface {
	Get(string) any
	GetChange(string) (any, any)
	GetOk(string) (any, bool)
	GetRawConfig() cty.Value
	GetRawPlan() cty.Value
	GetRawState() cty.Value
	HasChange(string) bool
	HasChanges(...string) bool
	Id() string
}

// ResourceForceNewDiffer is a ResourceDiffer that can also force resource
// replacement. Only *schema.ResourceDiff (i.e., the value passed to
// CustomizeDiff functions) implements this; *schema.ResourceData does not.
type ResourceForceNewDiffer interface {
	ResourceDiffer
	ForceNew(key string) error
}

// ForceNewIfChanged calls ForceNew on the given key only when the underlying
// diff records a change for it. This avoids the "ForceNew: No changes for X"
// error from the SDK during apply expansion when a previously-unknown
// configured value resolves to the same value already in state.
func ForceNewIfChanged(diff ResourceForceNewDiffer, key string) error {
	if !diff.HasChange(key) {
		return nil
	}
	return diff.ForceNew(key)
}

// AnyValues returns whether or not any of the given keys has a value.
func AnyValues(d ResourceDiffer, keys ...string) bool {
	for _, key := range keys {
		if _, ok := d.GetOk(key); ok {
			return true
		}
	}
	return false
}
