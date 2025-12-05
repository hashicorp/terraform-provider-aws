// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

// Exports for use in tests only.
var (
	ResourceDBCluster  = newDBClusterResource
	ResourceDBInstance = newDBInstanceResource

	FindDBClusterByID  = findDBClusterByID
	FindDBInstanceByID = findDBInstanceByID
)
