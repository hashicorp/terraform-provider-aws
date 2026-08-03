// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package billing

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_billing_service_account", name="Service Account")
func newServiceAccountDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &billingServiceAccountDataSource{}

	return d, nil
}

type billingServiceAccountDataSource struct {
	framework.DataSourceWithModel[billingServiceAccountDataSourceModel]
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
	const serviceAccountID = "386209384616"
	data.ARN = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().GlobalARNWithAccount(ctx, "iam", serviceAccountID, "root"))
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, serviceAccountID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type billingServiceAccountDataSourceModel struct {
	ARN types.String `tfsdk:"arn"`
	ID  types.String `tfsdk:"id"`
}
