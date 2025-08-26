// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import "context"

type ServiceTagUpdater interface {
	UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error
}

type ResourceTypeTagUpdater interface {
	UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error
}
