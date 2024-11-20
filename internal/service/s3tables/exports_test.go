// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	NewResourceNamespace         = newResourceNamespace
	NewResourceTable             = newResourceTable
	NewResourceTableBucket       = newResourceTableBucket
	NewResourceTableBucketPolicy = newResourceTableBucketPolicy

	FindNamespace         = findNamespace
	FindTable             = findTable
	FindTableBucket       = findTableBucket
	FindTableBucketPolicy = findTableBucketPolicy

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
