// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// withMetadata is intended to be embedded in all resources and data sources.
type withMetadata struct{}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (*withMetadata) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method is implemented in the wrappers.
	panic("not implemented") // lintignore:R009
}
