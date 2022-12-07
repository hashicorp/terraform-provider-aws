package sts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourceCallerIdentity)
}

// newDataSourceCallerIdentity instantiates a new DataSource for the aws_caller_identity data source.
func newDataSourceCallerIdentity(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCallerIdentity{}, nil
}

type dataSourceCallerIdentity struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceCallerIdentity) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_caller_identity"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceCallerIdentity) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"account_id": {
				Type:     types.StringType,
				Computed: true,
			},
			"arn": {
				Type:     types.StringType,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"user_id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}

	return schema, nil
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceCallerIdentity) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceCallerIdentityData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().STSConn

	output, err := FindCallerIdentity(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading STS Caller Identity", err.Error())

		return
	}

	accountID := aws.StringValue(output.Account)
	data.AccountID = types.StringValue(accountID)
	data.ARN = flex.StringToFrameworkLegacy(ctx, output.Arn)
	data.ID = types.StringValue(accountID)
	data.UserID = flex.StringToFrameworkLegacy(ctx, output.UserId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceCallerIdentityData struct {
	AccountID types.String `tfsdk:"account_id"`
	ARN       types.String `tfsdk:"arn"`
	ID        types.String `tfsdk:"id"`
	UserID    types.String `tfsdk:"user_id"`
}
