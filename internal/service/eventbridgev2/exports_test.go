// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

var (
	ResourceEventBus       = newEventBusResource
	ResourceEventSource    = newEventSourceResource
	ResourceResourcePolicy = newResourcePolicyResource
	ResourceSubscriber     = newSubscriberResource

	FindEventBusByARN       = findEventBusByARN
	FindEventSourceByARN    = findEventSourceByARN
	FindResourcePolicyByARN = findResourcePolicyByARN
	FindSubscriberByARN     = findSubscriberByARN
)
