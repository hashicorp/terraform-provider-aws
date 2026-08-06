// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	ResourceNamespace                       = newNamespaceResource
	ResourceTable                           = newTableResource
	ResourceTableBucket                     = newTableBucketResource
	ResourceTableBucketMetricsConfiguration = newTableBucketMetricsConfigurationResource
	ResourceTableBucketPolicy               = newTableBucketPolicyResource
	ResourceTableBucketReplication          = newTableBucketReplicationResource
	ResourceTablePolicy                     = newTablePolicyResource
	ResourceTableReplication                = newTableReplicationResource

	FindNamespaceByTwoPartKey                = findNamespaceByTwoPartKey
	FindTableByThreePartKey                  = findTableByThreePartKey
	FindTableBucketByARN                     = findTableBucketByARN
	FindTableBucketMetricsConfigurationByARN = findTableBucketMetricsConfigurationByARN
	FindTableBucketPolicyByARN               = findTableBucketPolicyByARN
	FindTableBucketReplicationByARN          = findTableBucketReplicationByARN
	FindTablePolicyByThreePartKey            = findTablePolicyByThreePartKey
	FindTableReplicationByARN                = findTableReplicationByARN

	TableIDFromTableARN = tableIDFromTableARN
)

type (
	TableIdentifier = tableIdentifier
)
