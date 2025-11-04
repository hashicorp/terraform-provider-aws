// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// ResourceWithConfigure is a structure to be embedded within a Resource that implements the ResourceWithConfigure interface.
type ResourceWithConfigure struct {
	withMeta
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (*ResourceWithConfigure) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method is implemented in the wrappers.
	panic("not implemented") // lintignore:R009
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *ResourceWithConfigure) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}
