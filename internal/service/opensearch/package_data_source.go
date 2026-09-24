// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearch_package", name="Package")
func newPackageDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &packageDataSource{}, nil
}

type packageDataSource struct {
	framework.DataSourceWithModel[packageDataSourceModel]
}

func (d *packageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"available_package_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrEngineVersion: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"package_description": schema.StringAttribute{
				Computed: true,
			},
			"package_id": schema.StringAttribute{
				Computed: true,
			},
			"package_name": schema.StringAttribute{
				Required: true,
			},
			"package_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PackageType](),
				Computed:   true,
			},
		},
	}
}

func (d *packageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data packageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().OpenSearchClient(ctx)

	pkg, err := findPackageByName(ctx, conn, data.PackageName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearch, create.ErrActionReading, "Package", data.PackageName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, pkg, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = data.PackageID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type packageDataSourceModel struct {
	framework.WithRegionModel
	AvailablePackageVersion types.String                             `tfsdk:"available_package_version"`
	EngineVersion           types.String                             `tfsdk:"engine_version"`
	ID                      types.String                             `tfsdk:"id"`
	PackageDescription      types.String                             `tfsdk:"package_description"`
	PackageID               types.String                             `tfsdk:"package_id"`
	PackageName             types.String                             `tfsdk:"package_name"`
	PackageType             fwtypes.StringEnum[awstypes.PackageType] `tfsdk:"package_type"`
}

func findPackageByName(ctx context.Context, conn *opensearch.Client, name string) (*awstypes.PackageDetails, error) {
	input := &opensearch.DescribePackagesInput{
		Filters: []awstypes.DescribePackagesFilter{
			{
				Name:  awstypes.DescribePackagesFilterNamePackageName,
				Value: []string{name},
			},
		},
	}

	return findPackage(ctx, conn, input)
}
