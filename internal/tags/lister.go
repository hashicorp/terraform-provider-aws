// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import "context"

type ServiceTagLister interface {
	ListTags(ctx context.Context, meta any, identifier string) error
}

type ResourceTypeTagLister interface {
	ListTags(ctx context.Context, meta any, identifier, resourceType string) error
}
