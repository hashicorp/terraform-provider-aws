// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	NewResourceNamespace         = newResourceNamespace
	NewResourceTableBucket       = newResourceTableBucket
	NewResourceTableBucketPolicy = newResourceTableBucketPolicy

	FindNamespace         = findNamespace
	FindTableBucket       = findTableBucket
	FindTableBucketPolicy = findTableBucketPolicy
)

const (
	ResNameNamespace   = resNameNamespace
	ResNameTableBucket = resNameTableBucket

	NamespaceIDSeparator = namespaceIDSeparator
)
