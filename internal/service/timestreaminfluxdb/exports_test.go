// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

// Exports for use in tests only.
var (
	ResourceDBCluster  = newResourceDBCluster
	ResourceDBInstance = newResourceDBInstance

	FindDBClusterByID  = findDBClusterByID
	FindDBInstanceByID = findDBInstanceByID
)
