// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_db_proxies",name=Proxies)
func newDBProxiesDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dbProxiesDataSource{}, nil
}

type dbProxiesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *dbProxiesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_db_proxies"
}

func (d *dbProxiesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"proxy_arns": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrNames: schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dbProxiesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data proxiesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().RDSClient(ctx)

	// It is not possible to filter by the proxy parameters, for future use
	input := &rds.DescribeDBProxiesInput{}

	proxies, err := findDBProxies(ctx, conn, input, tfslices.PredicateTrue[*rdsTypes.DBProxy]())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading RDS Proxies: %s"), err.Error())
		return
	}

	var proxyARNs []string
	var proxyNames []string

	for _, proxy := range proxies {
		proxyARNs = append(proxyARNs, aws.ToString(proxy.DBProxyArn))
		proxyNames = append(proxyNames, aws.ToString(proxy.DBProxyName))
	}

	data.Names = fwflex.FlattenFrameworkStringValueSetLegacy(ctx, proxyNames)
	data.ProxyARNs = fwflex.FlattenFrameworkStringValueSetLegacy(ctx, proxyARNs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type proxiesDataSourceModel struct {
	ProxyARNs types.Set `tfsdk:"proxy_arns"`
	Names     types.Set `tfsdk:"proxy_names"`
}
