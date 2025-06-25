// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	NewResourceNamespace         = newNamespaceResource
	NewResourceTable             = newTableResource
	NewResourceTableBucket       = newTableBucketResource
	NewResourceTableBucketPolicy = newTableBucketPolicyResource
	ResourceTablePolicy          = newTablePolicyResource

	FindNamespace         = findNamespace
	FindTable             = findTable
	FindTableBucket       = findTableBucket
	FindTableBucketPolicy = findTableBucketPolicy
	FindTablePolicy       = findTablePolicy

	TableIDFromTableARN = tableIDFromTableARN
)

const (
	ResNameNamespace   = resNameNamespace
	ResNameTableBucket = resNameTableBucket

	NamespaceIDSeparator = namespaceIDSeparator
)

type (
	TableIdentifier = tableIdentifier
)
