// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

var (
	NewResourceNamespace   = newResourceNamespace
	NewResourceTableBucket = newResourceTableBucket

	FindNamespace   = findNamespace
	FindTableBucket = findTableBucket
)

const (
	ResNameNamespace   = resNameNamespace
	ResNameTableBucket = resNameTableBucket

	NamespaceIDSeparator = namespaceIDSeparator
)
