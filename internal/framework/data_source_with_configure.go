// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// DataSourceWithConfigure is a structure to be embedded within a DataSource that implements the DataSourceWithConfigure interface.
type DataSourceWithConfigure struct {
	withMeta
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceWithConfigure) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

type dataSourceReader[T any] interface {
	// Read is called when the provider must read data source values in order to update state.
	// On entry `data` contains Config values and on return `data` is written to State.
	OnRead(ctx context.Context, data *T) diag.Diagnostics
}

type DataSourceWithConfigureEx[T any] struct {
	DataSourceWithConfigure
	impl dataSourceReader[T]
}

type dataSource[T any, U any] interface {
	datasource.DataSourceWithConfigure
	dataSourceReader[U]
	setImpl(dataSourceReader[U])
	*T
}

func NewDataSource[T any, U any, V dataSource[T, U]]() datasource.DataSourceWithConfigure {
	var v V = new(T)
	v.setImpl(v)
	return v
}

// SetImpl sets the reader implementation.
func (d *DataSourceWithConfigureEx[T]) setImpl(impl dataSourceReader[T]) {
	d.impl = impl
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *DataSourceWithConfigureEx[T]) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data T

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(d.impl.OnRead(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}
