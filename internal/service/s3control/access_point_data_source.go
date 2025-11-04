// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3_access_point", name="Access Point")
// @Tags(identifierAttribute="arn")
func newAccessPointDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &accessPointDataSource{}, nil
}

type accessPointDataSource struct {
	framework.DataSourceWithModel[accessPointDataSourceModel]
}

func (d *accessPointDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			names.AttrAlias: schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrBucket: schema.StringAttribute{
				Computed: true,
			},
			"bucket_account_id": schema.StringAttribute{
				Computed: true,
			},
			"data_source_id": schema.StringAttribute{
				Computed: true,
			},
			"data_source_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrEndpoints: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"network_origin": schema.StringAttribute{
				Computed: true,
			},
			"public_access_block_configuration": framework.DataSourceComputedListOfObjectAttribute[publicAccessBlockConfigurationModel](ctx),
			names.AttrTags:                      tftags.TagsAttributeComputedOnly(),
			names.AttrVPCConfiguration:          framework.DataSourceComputedListOfObjectAttribute[vpcConfigurationModel](ctx),
		},
	}
}

func (d *accessPointDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data accessPointDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3ControlClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	accountID := fwflex.StringValueFromFramework(ctx, data.AccountID)
	if accountID == "" {
		accountID = d.Meta().AccountID(ctx)
	}
	output, err := findAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Point (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.AccountID = fwflex.StringValueToFramework(ctx, accountID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type accessPointDataSourceModel struct {
	framework.WithRegionModel
	AccessPointARN                 types.String                                                         `tfsdk:"arn"`
	AccountID                      types.String                                                         `tfsdk:"account_id"`
	Alias                          types.String                                                         `tfsdk:"alias"`
	Bucket                         types.String                                                         `tfsdk:"bucket"`
	BucketAccountID                types.String                                                         `tfsdk:"bucket_account_id"`
	DataSourceID                   types.String                                                         `tfsdk:"data_source_id"`
	DataSourceType                 types.String                                                         `tfsdk:"data_source_type"`
	Endpoints                      fwtypes.MapOfString                                                  `tfsdk:"endpoints"`
	Name                           types.String                                                         `tfsdk:"name"`
	NetworkOrigin                  types.String                                                         `tfsdk:"network_origin"`
	PublicAccessBlockConfiguration fwtypes.ListNestedObjectValueOf[publicAccessBlockConfigurationModel] `tfsdk:"public_access_block_configuration"`
	Tags                           tftags.Map                                                           `tfsdk:"tags"`
	VPCConfiguration               fwtypes.ListNestedObjectValueOf[vpcConfigurationModel]               `tfsdk:"vpc_configuration"`
}

type publicAccessBlockConfigurationModel struct {
	BlockPublicACLs       types.Bool `tfsdk:"block_public_acls"`
	BlockPublicPolicy     types.Bool `tfsdk:"block_public_policy"`
	IgnorePublicACLs      types.Bool `tfsdk:"ignore_public_acls"`
	RestrictPublicBuckets types.Bool `tfsdk:"restrict_public_buckets"`
}

type vpcConfigurationModel struct {
	VpcID types.String `tfsdk:"vpc_id"`
}
