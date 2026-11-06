// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

var (
	ResourceEventBus    = newEventBusResource
	ResourceEventSource = newEventSourceResource

	FindEventBusByARN    = findEventBusByARN
	FindEventSourceByARN = findEventSourceByARN
)
