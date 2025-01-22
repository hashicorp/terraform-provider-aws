// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

// ResourceDiffer exposes the interface for accessing changes in a resource
// Implementations:
// * schema.ResourceData
// * schema.ResourceDiff
// Matches the public part of helper/schema/resourceDiffer:
// https://github.com/hashicorp/terraform-plugin-sdk/blob/28e631776d97f0a5a5942b3524814addbef90875/helper/schema/schema.go#L1104-L1112
type ResourceDiffer interface {
	Get(string) interface{}
	GetChange(string) (interface{}, interface{})
	GetOk(string) (interface{}, bool)
	HasChange(string) bool
	HasChanges(...string) bool
	Id() string
}

// HasNonZeroValues returns true if any of the keys have non-zero values.
func HasNonZeroValues(d ResourceDiffer, keys ...string) bool {
	for _, key := range keys {
		if _, ok := d.GetOk(key); ok {
			return true
		}
	}
	return false
}
