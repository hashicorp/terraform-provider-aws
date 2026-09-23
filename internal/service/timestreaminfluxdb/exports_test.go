// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

// Exports for use in tests only.
var (
	ResourceDBBackup   = newDBBackupResource
	ResourceDBCluster  = newDBClusterResource
	ResourceDBInstance = newDBInstanceResource

	FindDBBackupByID   = findDBBackupByID
	FindDBClusterByID  = findDBClusterByID
	FindDBInstanceByID = findDBInstanceByID
)
