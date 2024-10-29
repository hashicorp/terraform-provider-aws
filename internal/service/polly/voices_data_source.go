// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package polly

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/polly"
	awstypes "github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Voices")
func newDataSourceVoices(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceVoices{}, nil
}

const (
	DSNameVoices = "Voices Data Source"
)

type dataSourceVoices struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceVoices) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_polly_voices"
}

func (d *dataSourceVoices) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEngine: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Engine](),
				Optional:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"include_additional_language_codes": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrLanguageCode: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LanguageCode](),
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"voices": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[voicesData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"additional_language_codes": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Computed:    true,
						},
						"gender": schema.StringAttribute{
							Computed: true,
						},
						names.AttrID: schema.StringAttribute{
							Computed: true,
						},
						names.AttrLanguageCode: schema.StringAttribute{
							Computed: true,
						},
						"language_name": schema.StringAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"supported_engines": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceVoices) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().PollyClient(ctx)

	var data dataSourceVoicesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(d.Meta().AccountID)

	input := &polly.DescribeVoicesInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No paginator helper so pagination must be done manually
	out := &polly.DescribeVoicesOutput{}
	for {
		page, err := conn.DescribeVoices(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Polly, create.ErrActionReading, DSNameVoices, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if page == nil {
			break
		}

		if len(page.Voices) > 0 {
			out.Voices = append(out.Voices, page.Voices...)
		}

		input.NextToken = page.NextToken
		if page.NextToken == nil {
			break
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceVoicesData struct {
	Engine                         fwtypes.StringEnum[awstypes.Engine]         `tfsdk:"engine"`
	ID                             types.String                                `tfsdk:"id"`
	IncludeAdditionalLanguageCodes types.Bool                                  `tfsdk:"include_additional_language_codes"`
	LanguageCode                   fwtypes.StringEnum[awstypes.LanguageCode]   `tfsdk:"language_code"`
	Voices                         fwtypes.ListNestedObjectValueOf[voicesData] `tfsdk:"voices"`
}

type voicesData struct {
	AdditionalLanguageCodes fwtypes.ListValueOf[types.String] `tfsdk:"additional_language_codes"`
	Gender                  types.String                      `tfsdk:"gender"`
	ID                      types.String                      `tfsdk:"id"`
	LanguageCode            types.String                      `tfsdk:"language_code"`
	LanguageName            types.String                      `tfsdk:"language_name"`
	Name                    types.String                      `tfsdk:"name"`
	SupportedEngines        fwtypes.ListValueOf[types.String] `tfsdk:"supported_engines"`
}
