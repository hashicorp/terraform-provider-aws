// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	ResourceNamespace              = newNamespaceResource
	ResourceTable                  = newTableResource
	ResourceTableBucket            = newTableBucketResource
	ResourceTableBucketPolicy      = newTableBucketPolicyResource
	ResourceTableBucketReplication = newTableBucketReplicationResource
	ResourceTablePolicy            = newTablePolicyResource
	ResourceTableReplication       = newTableReplicationResource

	FindNamespaceByTwoPartKey       = findNamespaceByTwoPartKey
	FindTableByThreePartKey         = findTableByThreePartKey
	FindTableBucketByARN            = findTableBucketByARN
	FindTableBucketPolicyByARN      = findTableBucketPolicyByARN
	FindTableBucketReplicationByARN = findTableBucketReplicationByARN
	FindTablePolicyByThreePartKey   = findTablePolicyByThreePartKey
	FindTableReplicationByARN       = findTableReplicationByARN

	TableIDFromTableARN = tableIDFromTableARN
)

type (
	TableIdentifier = tableIdentifier
)
