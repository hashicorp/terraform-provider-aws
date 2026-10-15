// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_bedrock_use_case_for_model_access", name="Use Case For Model Access")
// @Region(global=true)
func newUseCaseForModelAccessDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &useCaseForModelAccessDataSource{}, nil
}

const (
	DSNameUseCaseForModelAccess = "Use Case For Model Access Data Source"
)

type useCaseForModelAccessDataSource struct {
	framework.DataSourceWithModel[useCaseForModelAccessDataSourceModel]
}

func (d *useCaseForModelAccessDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"form_data": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *useCaseForModelAccessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BedrockClient(ctx)

	var data useCaseForModelAccessDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.GetUseCaseForModelAccessInput{}
	out, err := conn.GetUseCaseForModelAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	if out == nil || out.FormData == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("Empty Response"))
		return
	}

	v, diags := flattenFormData(ctx, out.FormData)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	data.FormData = v

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type useCaseForModelAccessDataSourceModel struct {
	FormData types.String `tfsdk:"form_data"`
}
