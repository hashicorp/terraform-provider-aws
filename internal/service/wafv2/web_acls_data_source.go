// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_wafv2_web_acls", name="Web ACLs")
func newDataSourceWebACLs(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceWebACLs{}, nil
}

const (
	DSNameWebACLs = "Web ACLs Data Source"
)

type dataSourceWebACLs struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceWebACLs) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrNames: schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			names.AttrScope: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.Scope](),
				},
			},
		},
	}
}

func (d *dataSourceWebACLs) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().WAFV2Client(ctx)

	var data dataSourceWebACLsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := ListWebACLs(ctx, conn, data.Scope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, DSNameWebACLs, data.Scope.String(), err),
			err.Error(),
		)
		return
	}

	webACLNames := make([]string, 0, len(out))
	for _, webACL := range out {
		webACLNames = append(webACLNames, *webACL.Name)
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, struct {
		Names []string
		Scope string
	}{webACLNames, data.Scope.ValueString()}, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ListWebACLs(ctx context.Context, conn *wafv2.Client, scope string) ([]awstypes.WebACLSummary, error) {
	input := &wafv2.ListWebACLsInput{
		Scope: awstypes.Scope(scope),
	}

	var output []awstypes.WebACLSummary
	err := listWebACLsPages(ctx, conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
		if page == nil || page.WebACLs == nil {
			return !lastPage
		}
		output = append(output, page.WebACLs...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

type dataSourceWebACLsModel struct {
	Names fwtypes.ListValueOf[types.String] `tfsdk:"names"`
	Scope types.String                      `tfsdk:"scope"`
}
