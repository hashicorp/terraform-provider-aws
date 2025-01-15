// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_billing_service_account", name="Billing Service Account")
func newBillingServiceAccountDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &billingServiceAccountDataSource{}

	return d, nil
}

type billingServiceAccountDataSource struct {
	framework.DataSourceWithConfigure
}

func (*billingServiceAccountDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_billing_service_account"
}

func (d *billingServiceAccountDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *billingServiceAccountDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data billingServiceAccountDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// See http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-getting-started.html#step-2
	const billingAccountID = "386209384616"

	arn := arn.ARN{
		Partition: d.Meta().Partition(ctx),
		Service:   "iam",
		AccountID: billingAccountID,
		Resource:  "root",
	}
	data.ARN = fwflex.StringValueToFrameworkLegacy(ctx, arn.String())
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, billingAccountID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type billingServiceAccountDataSourceModel struct {
	ARN types.String `tfsdk:"arn"`
	ID  types.String `tfsdk:"id"`
}
