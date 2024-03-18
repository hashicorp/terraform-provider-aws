// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

// Exports for use in tests only.
var (
	ResourceBroker        = resourceBroker
	ResourceConfiguration = resourceConfiguration

	FindBrokerByID        = findBrokerByID
	FindConfigurationByID = findConfigurationByID

	WaitBrokerRebooted = waitBrokerRebooted
	WaitBrokerDeleted  = waitBrokerDeleted
)
