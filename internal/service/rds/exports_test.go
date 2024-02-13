// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

// Exports for use in tests only.
var (
	ResourceProxy = resourceProxy

	FindDBInstanceByID               = findDBInstanceByIDSDKv1
	FindDBProxyByName                = findDBProxyByName
	ListTags                         = listTags
	NewBlueGreenOrchestrator         = newBlueGreenOrchestrator
	ParseDBInstanceARN               = parseDBInstanceARN
	WaitBlueGreenDeploymentDeleted   = waitBlueGreenDeploymentDeleted
	WaitBlueGreenDeploymentAvailable = waitBlueGreenDeploymentAvailable
	WaitDBInstanceAvailable          = waitDBInstanceAvailableSDKv2
	WaitDBInstanceDeleted            = waitDBInstanceDeleted

	ErrCodeInvalidParameterCombination = errCodeInvalidParameterCombination
	ErrCodeInvalidParameterValue       = errCodeInvalidParameterValue
)
