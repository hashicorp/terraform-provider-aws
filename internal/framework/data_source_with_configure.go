// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// DataSourceWithConfigure is a structure to be embedded within a DataSource that implements the DataSourceWithConfigure interface.
type DataSourceWithConfigure struct {
	withMeta
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (*DataSourceWithConfigure) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	// This method is implemented in the wrappers.
	panic("not implemented") // lintignore:R009
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceWithConfigure) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}
