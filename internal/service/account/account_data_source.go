// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_account_account", name="Account")
func newAccountDataSource(context.Context) (datasource.DataSourceWithConfigure, error) { // nosemgrep:ci.account-in-func-name
	return &dataSourceAccount{}, nil
}

type dataSourceAccount struct {
	framework.DataSourceWithModel[dataSourceAccountModel]
}

func (d *dataSourceAccount) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_created_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrAccountID: schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"account_name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AccountClient(ctx)

	// TIP: -- 2. Fetch the config
	var data dataSourceAccountModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccountInformation(ctx, conn, data.AccountID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Account")), smerr.ID)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID)
}

type dataSourceAccountModel struct {
	AccountCreatedDate timetypes.RFC3339 `tfsdk:"account_created_date"`
	AccountID          types.String      `tfsdk:"account_id"`
	AccountName        types.String      `tfsdk:"account_name"`
}

func findAccountInformation(ctx context.Context, conn *account.Client, accountID string) (*account.GetAccountInformationOutput, error) { // nosemgrep:ci.account-in-func-name
	input := account.GetAccountInformationInput{}
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetAccountInformation(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
