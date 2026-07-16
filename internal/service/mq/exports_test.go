// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mq

// Exports for use in tests only.
var (
	ResourceBroker        = resourceBroker
	ResourceConfiguration = resourceConfiguration

	DiffBrokerUsers             = diffBrokerUsers
	FindBrokerByID              = findBrokerByID
	FindConfigurationByID       = findConfigurationByID
	FlattenResourceShareARNs    = flattenResourceShareARNs
	NormalizeEngineVersion      = normalizeEngineVersion
	SortBrokerInstanceEndpoints = sortBrokerInstanceEndpoints
	ValidateBrokerName          = validateBrokerName
	ValidBrokerPassword         = validBrokerPassword
	WaitBrokerRebooted          = waitBrokerRebooted
	WaitBrokerDeleted           = waitBrokerDeleted
)
