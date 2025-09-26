// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	ResourceNamespace         = newNamespaceResource
	ResourceTable             = newTableResource
	ResourceTableBucket       = newTableBucketResource
	ResourceTableBucketPolicy = newTableBucketPolicyResource
	ResourceTablePolicy       = newTablePolicyResource

	FindNamespaceByTwoPartKey     = findNamespaceByTwoPartKey
	FindTableByThreePartKey       = findTableByThreePartKey
	FindTableBucketByARN          = findTableBucketByARN
	FindTableBucketPolicyByARN    = findTableBucketPolicyByARN
	FindTablePolicyByThreePartKey = findTablePolicyByThreePartKey

	TableIDFromTableARN = tableIDFromTableARN
)

type (
	TableIdentifier = tableIdentifier
)
