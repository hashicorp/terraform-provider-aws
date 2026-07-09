// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_guardrails", name="Guardrails")
func newGuardrailsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &guardrailsDataSource{}, nil
}

type guardrailsDataSource struct {
	framework.DataSourceWithModel[guardrailsDataSourceModel]
}

func (d *guardrailsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"guardrail_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^(([a-z0-9]+)|(arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:[0-9]{12}:guardrail/[a-z0-9]+))$`),
						"must be a guardrail ID (lowercase alphanumeric) or a full guardrail ARN",
					),
				},
			},
			"guardrails": framework.DataSourceComputedListOfObjectAttribute[guardrailSummaryModel](ctx),
		},
	}
}

func (d *guardrailsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data guardrailsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.ListGuardrailsInput{}
	if !data.GuardrailIdentifier.IsNull() {
		input.GuardrailIdentifier = data.GuardrailIdentifier.ValueStringPointer()
	}

	summaries, err := findGuardrails(ctx, conn, input)
	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Guardrails", err.Error())
		return
	}

	items := make([]guardrailSummaryModel, 0, len(summaries))
	for _, summary := range summaries {
		var item guardrailSummaryModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, &summary, &item)...)
		if response.Diagnostics.HasError() {
			return
		}

		tags, err := listTags(ctx, conn, aws.ToString(summary.Arn))
		if err != nil {
			response.Diagnostics.AddError(
				"listing tags for Bedrock Guardrail",
				fmt.Errorf("listing tags for Bedrock Guardrail (%s): %w", aws.ToString(summary.Arn), err).Error(),
			)
			return
		}
		item.Tags = fwflex.FlattenFrameworkStringValueMapOfString(ctx, tags.Map())

		items = append(items, item)
	}

	guardrails, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, items)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Guardrails = guardrails
	data.ID = types.StringValue(d.Meta().Region(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findGuardrails(ctx context.Context, conn *bedrock.Client, input *bedrock.ListGuardrailsInput) ([]awstypes.GuardrailSummary, error) {
	var output []awstypes.GuardrailSummary

	pages := bedrock.NewListGuardrailsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.Guardrails...)
	}

	return output, nil
}

type guardrailsDataSourceModel struct {
	framework.WithRegionModel
	GuardrailIdentifier types.String                                           `tfsdk:"guardrail_identifier"`
	Guardrails          fwtypes.ListNestedObjectValueOf[guardrailSummaryModel] `tfsdk:"guardrails"`
	ID                  types.String                                           `tfsdk:"id"`
}

type guardrailSummaryModel struct {
	Arn         fwtypes.ARN                                  `tfsdk:"arn"`
	CreatedAt   timetypes.RFC3339                            `tfsdk:"created_at"`
	Description types.String                                 `tfsdk:"description"`
	ID          types.String                                 `tfsdk:"guardrail_id"`
	Name        types.String                                 `tfsdk:"name"`
	Status      fwtypes.StringEnum[awstypes.GuardrailStatus] `tfsdk:"status"`
	Tags        fwtypes.MapOfString                          `tfsdk:"tags"`
	UpdatedAt   timetypes.RFC3339                            `tfsdk:"updated_at"`
	Version     types.String                                 `tfsdk:"version"`
}
