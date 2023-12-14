// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

// Exports for use in tests only.
var (
	FindDBInstanceByID = findDBInstanceByIDSDKv1

	ListTags = listTags

	NewBlueGreenOrchestrator = newBlueGreenOrchestrator

	WaitBlueGreenDeploymentDeleted   = waitBlueGreenDeploymentDeleted
	WaitBlueGreenDeploymentAvailable = waitBlueGreenDeploymentAvailable

	ParseDBInstanceARN = parseDBInstanceARN

	WaitDBInstanceAvailable = waitDBInstanceAvailableSDKv2
	WaitDBInstanceDeleted   = waitDBInstanceDeleted
)

const (
	ErrCodeInvalidParameterCombination = errCodeInvalidParameterCombination
	ErrCodeInvalidParameterValue       = errCodeInvalidParameterValue
)
