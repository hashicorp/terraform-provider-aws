// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"fmt"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwschema "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_guardrail", name="Guardrail")
func newGuardrailDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &guardrailDataSource{}, nil
}

type guardrailDataSource struct {
	framework.DataSourceWithModel[guardrailDataSourceModel]
}

func (d *guardrailDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"blocked_input_messaging": schema.StringAttribute{
				Computed: true,
			},
			"blocked_outputs_messaging": schema.StringAttribute{
				Computed: true,
			},
			"content_policy_config":              framework.DataSourceComputedListOfObjectAttribute[guardrailContentPolicyConfigModel](ctx),
			"contextual_grounding_policy_config": framework.DataSourceComputedListOfObjectAttribute[contextualGroundingPolicyConfig](ctx),
			"cross_region_config":                framework.DataSourceComputedListOfObjectAttribute[guardrailCrossRegionConfigModel](ctx),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"latest": schema.BoolAttribute{
				Optional: true,
				Validators: []fwschema.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot(names.AttrVersion)),
				},
			},
			"guardrail_id": schema.StringAttribute{
				Computed: true,
			},
			"guardrail_identifier": schema.StringAttribute{
				Required: true,
				Validators: []fwschema.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^(([a-z0-9]+)|(arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:[0-9]{12}:guardrail/[a-z0-9]+))$`),
						"must be a guardrail ID (lowercase alphanumeric) or a full guardrail ARN",
					),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"sensitive_information_policy_config": framework.DataSourceComputedListOfObjectAttribute[sensitiveInformationPolicyConfig](ctx),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GuardrailStatus](),
				Computed:   true,
			},
			names.AttrTags:        tftags.TagsAttributeComputedOnly(),
			"topic_policy_config": framework.DataSourceComputedListOfObjectAttribute[guardrailTopicPolicyConfigModel](ctx),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []fwschema.String{
					stringvalidator.ConflictsWith(path.MatchRoot("latest")),
				},
			},
			"word_policy_config": framework.DataSourceComputedListOfObjectAttribute[wordPolicyConfig](ctx),
		},
	}
}

func (d *guardrailDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data guardrailDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)
	id := data.GuardrailIdentifier.ValueString()

	var resolvedVersion string

	if data.FetchLatest.ValueBool() {
		summaries, err := findGuardrails(ctx, conn, &bedrock.ListGuardrailsInput{
			GuardrailIdentifier: aws.String(id),
		})
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("listing Bedrock Guardrail (%s) versions", id), err.Error())
			return
		}
		maxVersion := 0
		for _, s := range summaries {
			if aws.ToString(s.Version) == "DRAFT" {
				continue
			}
			if v, parseErr := strconv.Atoi(aws.ToString(s.Version)); parseErr == nil && v > maxVersion {
				maxVersion = v
			}
		}
		if maxVersion == 0 {
			response.Diagnostics.AddError(
				fmt.Sprintf("reading Bedrock Guardrail (%s)", id),
				"latest is true but no published versions exist; publish a version first",
			)
			return
		}
		resolvedVersion = strconv.Itoa(maxVersion)
	} else {
		resolvedVersion = data.Version.ValueString()
		if resolvedVersion == "" {
			resolvedVersion = "DRAFT"
		}
	}

	output, err := findGuardrailByTwoPartKey(ctx, conn, id, resolvedVersion)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Guardrail (%s)", id), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNameSuffix("Config"))...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manual mapping: API field KmsKeyArn does not match model field KmsKeyId.
	data.KmsKeyId = fwflex.StringToFramework(ctx, output.KmsKeyArn)

	// Manual mapping: API returns CrossRegionDetails but schema uses cross_region_config.
	if output.CrossRegionDetails != nil && output.CrossRegionDetails.GuardrailProfileArn != nil {
		cr := guardrailCrossRegionConfigModel{
			GuardrailProfileIdentifier: fwflex.StringToFrameworkARN(ctx, output.CrossRegionDetails.GuardrailProfileArn),
		}
		var crossRegionDiags diag.Diagnostics
		data.CrossRegionConfig, crossRegionDiags = fwtypes.NewListNestedObjectValueOfPtr(ctx, &cr)
		response.Diagnostics.Append(crossRegionDiags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	guardrailARN := aws.ToString(output.GuardrailArn)
	tags, err := listTags(ctx, conn, guardrailARN)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing tags for Bedrock Guardrail (%s)", guardrailARN), err.Error())
		return
	}
	data.Tags = tftags.FlattenStringValueMap(ctx, tags.IgnoreAWS().Map())
	data.ID = types.StringValue(guardrailARN)

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

type guardrailDataSourceModel struct {
	framework.WithRegionModel
	BlockedInputMessaging      types.String                                                       `tfsdk:"blocked_input_messaging"`
	BlockedOutputsMessaging    types.String                                                       `tfsdk:"blocked_outputs_messaging"`
	ContentPolicy              fwtypes.ListNestedObjectValueOf[guardrailContentPolicyConfigModel] `tfsdk:"content_policy_config"`
	ContextualGroundingPolicy  fwtypes.ListNestedObjectValueOf[contextualGroundingPolicyConfig]   `tfsdk:"contextual_grounding_policy_config"`
	CreatedAt                  timetypes.RFC3339                                                  `tfsdk:"created_at"`
	CrossRegionConfig          fwtypes.ListNestedObjectValueOf[guardrailCrossRegionConfigModel]   `tfsdk:"cross_region_config"`
	Description                types.String                                                       `tfsdk:"description"`
	FetchLatest                types.Bool                                                         `tfsdk:"latest"`
	GuardrailArn               fwtypes.ARN                                                        `tfsdk:"arn"`
	GuardrailIdentifier        types.String                                                       `tfsdk:"guardrail_identifier"`
	GuardrailID                types.String                                                       `tfsdk:"guardrail_id"`
	ID                         types.String                                                       `tfsdk:"id"`
	KmsKeyId                   types.String                                                       `tfsdk:"kms_key_arn"`
	Name                       types.String                                                       `tfsdk:"name"`
	SensitiveInformationPolicy fwtypes.ListNestedObjectValueOf[sensitiveInformationPolicyConfig]  `tfsdk:"sensitive_information_policy_config"`
	Status                     fwtypes.StringEnum[awstypes.GuardrailStatus]                       `tfsdk:"status"`
	Tags                       tftags.Map                                                         `tfsdk:"tags"`
	TopicPolicy                fwtypes.ListNestedObjectValueOf[guardrailTopicPolicyConfigModel]   `tfsdk:"topic_policy_config"`
	UpdatedAt                  timetypes.RFC3339                                                  `tfsdk:"updated_at"`
	Version                    types.String                                                       `tfsdk:"version"`
	WordPolicy                 fwtypes.ListNestedObjectValueOf[wordPolicyConfig]                  `tfsdk:"word_policy_config"`
}
