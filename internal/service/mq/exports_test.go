// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mq

// Exports for use in tests only.
var (
	ResourceBroker        = resourceBroker
	ResourceConfiguration = resourceConfiguration

	FindBrokerByID        = findBrokerByID
	FindConfigurationByID = findConfigurationByID

	NormalizeEngineVersion = normalizeEngineVersion

	WaitBrokerRebooted = waitBrokerRebooted
	WaitBrokerDeleted  = waitBrokerDeleted
)
