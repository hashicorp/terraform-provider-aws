// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_caller_identity", name="Caller Identity")
func newCallerIdentityDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &callerIdentityDataSource{}

	return d, nil
}

type callerIdentityDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *callerIdentityDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"user_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *callerIdentityDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data CallerIdentityDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().STSClient(ctx)

	output, err := findCallerIdentity(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading STS Caller Identity", err.Error())

		return
	}

	accountID := aws.ToString(output.Account)
	data.AccountID = types.StringValue(accountID)
	data.ARN = flex.StringToFrameworkLegacy(ctx, output.Arn)
	data.ID = types.StringValue(accountID)
	data.UserID = flex.StringToFrameworkLegacy(ctx, output.UserId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findCallerIdentity(ctx context.Context, conn *sts.Client) (*sts.GetCallerIdentityOutput, error) {
	input := &sts.GetCallerIdentityInput{}

	output, err := conn.GetCallerIdentity(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type CallerIdentityDataSourceModel struct {
	AccountID types.String `tfsdk:"account_id"`
	ARN       types.String `tfsdk:"arn"`
	ID        types.String `tfsdk:"id"`
	UserID    types.String `tfsdk:"user_id"`
}
