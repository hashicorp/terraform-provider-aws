// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

const (
	engineMemcached = "memcached"
	engineRedis     = "redis"
	engineValkey    = "valkey"
)

// engine_Values returns all elements of the Engine enum
func engine_Values() []string {
	return []string{
		engineMemcached,
		engineRedis,
		engineValkey,
	}
}

const (
	reservedCacheNodeStateActive         = "active"
	reservedCacheNodeStateRetired        = "retired"
	reservedCacheNodeStatePaymentPending = "payment-pending"
)
