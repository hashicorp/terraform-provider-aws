// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockruntime

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource("aws_bedrockruntime_apply_guardrail", name="Apply Guardrail")
func newApplyGuardrailEphemeralResource(context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &applyGuardrailEphemeralResource{}, nil
}

type applyGuardrailEphemeralResource struct {
	framework.EphemeralResourceWithModel[applyGuardrailEphemeralResourceModel]
}

func (e *applyGuardrailEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAction: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GuardrailAction](),
				Computed:   true,
			},
			"action_reason": schema.StringAttribute{
				Computed: true,
			},
			"assessments": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailAssessmentModel](ctx),
				Computed:   true,
				Sensitive:  true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[guardrailAssessmentModel](ctx),
				},
			},
			"guardrail_coverage": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailCoverageModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[guardrailCoverageModel](ctx),
				},
			},
			"guardrail_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
			},
			"guardrail_version": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 8),
				},
			},
			"output": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailOutputModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[guardrailOutputModel](ctx),
				},
			},
			names.AttrSource: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GuardrailContentSource](),
				Required:   true,
			},
			"usage": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailUsageModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[guardrailUsageModel](ctx),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"content": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailContentModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"text": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailTextModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"qualifiers": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringEnumType[awstypes.GuardrailContentQualifier](),
										Optional:   true,
									},
									"text": schema.StringAttribute{
										Required:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (e *applyGuardrailEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	conn := e.Meta().BedrockRuntimeClient(ctx)

	var data applyGuardrailEphemeralResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockruntime.ApplyGuardrailInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.ApplyGuardrail(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.GuardrailIdentifier.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.Result.Set(ctx, &data))
}

type applyGuardrailEphemeralResourceModel struct {
	framework.WithRegionModel
	Action              fwtypes.StringEnum[awstypes.GuardrailAction]              `tfsdk:"action"`
	ActionReason        types.String                                              `tfsdk:"action_reason"`
	Assessments         fwtypes.ListNestedObjectValueOf[guardrailAssessmentModel] `tfsdk:"assessments"`
	Content             fwtypes.ListNestedObjectValueOf[guardrailContentModel]    `tfsdk:"content"`
	GuardrailCoverage   fwtypes.ListNestedObjectValueOf[guardrailCoverageModel]   `tfsdk:"guardrail_coverage"`
	GuardrailIdentifier types.String                                              `tfsdk:"guardrail_identifier"`
	GuardrailVersion    types.String                                              `tfsdk:"guardrail_version"`
	Output              fwtypes.ListNestedObjectValueOf[guardrailOutputModel]     `tfsdk:"output"`
	Source              fwtypes.StringEnum[awstypes.GuardrailContentSource]       `tfsdk:"source"`
	Usage               fwtypes.ListNestedObjectValueOf[guardrailUsageModel]      `tfsdk:"usage"`
}

type guardrailContentModel struct {
	Text fwtypes.ListNestedObjectValueOf[guardrailTextModel] `tfsdk:"text"`
}

var _ flex.Expander = guardrailContentModel{}

func (m guardrailContentModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	text, diags := m.Text.ToPtr(ctx)
	if diags.HasError() {
		return nil, diags
	}
	if text == nil {
		smerr.AddError(ctx, &diags, errors.New("content must contain exactly one text block"))
		return nil, diags
	}

	var content awstypes.GuardrailContentBlockMemberText
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, text, &content.Value))
	if diags.HasError() {
		return nil, diags
	}

	return &content, diags
}

type guardrailTextModel struct {
	Qualifiers fwtypes.ListOfStringEnum[awstypes.GuardrailContentQualifier] `tfsdk:"qualifiers"`
	Text       types.String                                                 `tfsdk:"text"`
}

type guardrailOutputModel struct {
	Text types.String `tfsdk:"text"`
}

type guardrailCoverageModel struct {
	Images         fwtypes.ListNestedObjectValueOf[guardrailImageCoverageModel]          `tfsdk:"images"`
	TextCharacters fwtypes.ListNestedObjectValueOf[guardrailTextCharactersCoverageModel] `tfsdk:"text_characters"`
}

type guardrailImageCoverageModel struct {
	Guarded types.Int64 `tfsdk:"guarded"`
	Total   types.Int64 `tfsdk:"total"`
}

type guardrailTextCharactersCoverageModel struct {
	Guarded types.Int64 `tfsdk:"guarded"`
	Total   types.Int64 `tfsdk:"total"`
}

type guardrailUsageModel struct {
	AutomatedReasoningPolicies          types.Int64 `tfsdk:"automated_reasoning_policies"`
	AutomatedReasoningPolicyUnits       types.Int64 `tfsdk:"automated_reasoning_policy_units"`
	ContentPolicyImageUnits             types.Int64 `tfsdk:"content_policy_image_units"`
	ContentPolicyUnits                  types.Int64 `tfsdk:"content_policy_units"`
	ContextualGroundingPolicyUnits      types.Int64 `tfsdk:"contextual_grounding_policy_units"`
	SensitiveInformationPolicyFreeUnits types.Int64 `tfsdk:"sensitive_information_policy_free_units"`
	SensitiveInformationPolicyUnits     types.Int64 `tfsdk:"sensitive_information_policy_units"`
	TopicPolicyUnits                    types.Int64 `tfsdk:"topic_policy_units"`
	WordPolicyUnits                     types.Int64 `tfsdk:"word_policy_units"`
}

type guardrailAssessmentModel struct {
	AppliedGuardrailDetails    fwtypes.ListNestedObjectValueOf[appliedGuardrailDetailsModel]                       `tfsdk:"applied_guardrail_details"`
	AutomatedReasoningPolicy   fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningPolicyAssessmentModel]   `tfsdk:"automated_reasoning_policy"`
	ContentPolicy              fwtypes.ListNestedObjectValueOf[guardrailContentPolicyAssessmentModel]              `tfsdk:"content_policy"`
	ContextualGroundingPolicy  fwtypes.ListNestedObjectValueOf[guardrailContextualGroundingPolicyAssessmentModel]  `tfsdk:"contextual_grounding_policy"`
	InvocationMetrics          fwtypes.ListNestedObjectValueOf[guardrailInvocationMetricsModel]                    `tfsdk:"invocation_metrics"`
	SensitiveInformationPolicy fwtypes.ListNestedObjectValueOf[guardrailSensitiveInformationPolicyAssessmentModel] `tfsdk:"sensitive_information_policy"`
	TopicPolicy                fwtypes.ListNestedObjectValueOf[guardrailTopicPolicyAssessmentModel]                `tfsdk:"topic_policy"`
	WordPolicy                 fwtypes.ListNestedObjectValueOf[guardrailWordPolicyAssessmentModel]                 `tfsdk:"word_policy"`
}

type appliedGuardrailDetailsModel struct {
	GuardrailARN       types.String                                       `tfsdk:"guardrail_arn"`
	GuardrailID        types.String                                       `tfsdk:"guardrail_id"`
	GuardrailOrigin    fwtypes.ListOfStringEnum[awstypes.GuardrailOrigin] `tfsdk:"guardrail_origin"`
	GuardrailOwnership fwtypes.StringEnum[awstypes.GuardrailOwnership]    `tfsdk:"guardrail_ownership"`
	GuardrailVersion   types.String                                       `tfsdk:"guardrail_version"`
}

type guardrailContentPolicyAssessmentModel struct {
	Filters fwtypes.ListNestedObjectValueOf[guardrailContentFilterModel] `tfsdk:"filters"`
}

type guardrailContentFilterModel struct {
	Action         fwtypes.StringEnum[awstypes.GuardrailContentPolicyAction]     `tfsdk:"action"`
	Confidence     fwtypes.StringEnum[awstypes.GuardrailContentFilterConfidence] `tfsdk:"confidence"`
	Detected       types.Bool                                                    `tfsdk:"detected"`
	FilterStrength fwtypes.StringEnum[awstypes.GuardrailContentFilterStrength]   `tfsdk:"filter_strength"`
	Type           fwtypes.StringEnum[awstypes.GuardrailContentFilterType]       `tfsdk:"type"`
}

type guardrailContextualGroundingPolicyAssessmentModel struct {
	Filters fwtypes.ListNestedObjectValueOf[guardrailContextualGroundingFilterModel] `tfsdk:"filters"`
}

type guardrailContextualGroundingFilterModel struct {
	Action    fwtypes.StringEnum[awstypes.GuardrailContextualGroundingPolicyAction] `tfsdk:"action"`
	Detected  types.Bool                                                            `tfsdk:"detected"`
	Score     types.Float64                                                         `tfsdk:"score"`
	Threshold types.Float64                                                         `tfsdk:"threshold"`
	Type      fwtypes.StringEnum[awstypes.GuardrailContextualGroundingFilterType]   `tfsdk:"type"`
}

type guardrailInvocationMetricsModel struct {
	GuardrailCoverage          fwtypes.ListNestedObjectValueOf[guardrailCoverageModel] `tfsdk:"guardrail_coverage"`
	GuardrailProcessingLatency types.Int64                                             `tfsdk:"guardrail_processing_latency"`
	Usage                      fwtypes.ListNestedObjectValueOf[guardrailUsageModel]    `tfsdk:"usage"`
}

type guardrailSensitiveInformationPolicyAssessmentModel struct {
	PiiEntities fwtypes.ListNestedObjectValueOf[guardrailPiiEntityFilterModel] `tfsdk:"pii_entities"`
	Regexes     fwtypes.ListNestedObjectValueOf[guardrailRegexFilterModel]     `tfsdk:"regexes"`
}

type guardrailPiiEntityFilterModel struct {
	Action   fwtypes.StringEnum[awstypes.GuardrailSensitiveInformationPolicyAction] `tfsdk:"action"`
	Detected types.Bool                                                             `tfsdk:"detected"`
	Match    types.String                                                           `tfsdk:"match"`
	Type     fwtypes.StringEnum[awstypes.GuardrailPiiEntityType]                    `tfsdk:"type"`
}

type guardrailRegexFilterModel struct {
	Action   fwtypes.StringEnum[awstypes.GuardrailSensitiveInformationPolicyAction] `tfsdk:"action"`
	Detected types.Bool                                                             `tfsdk:"detected"`
	Match    types.String                                                           `tfsdk:"match"`
	Name     types.String                                                           `tfsdk:"name"`
	Regex    types.String                                                           `tfsdk:"regex"`
}

type guardrailTopicPolicyAssessmentModel struct {
	Topics fwtypes.ListNestedObjectValueOf[guardrailTopicModel] `tfsdk:"topics"`
}

type guardrailTopicModel struct {
	Action   fwtypes.StringEnum[awstypes.GuardrailTopicPolicyAction] `tfsdk:"action"`
	Detected types.Bool                                              `tfsdk:"detected"`
	Name     types.String                                            `tfsdk:"name"`
	Type     fwtypes.StringEnum[awstypes.GuardrailTopicType]         `tfsdk:"type"`
}

type guardrailWordPolicyAssessmentModel struct {
	CustomWords      fwtypes.ListNestedObjectValueOf[guardrailCustomWordModel]  `tfsdk:"custom_words"`
	ManagedWordLists fwtypes.ListNestedObjectValueOf[guardrailManagedWordModel] `tfsdk:"managed_word_lists"`
}

type guardrailCustomWordModel struct {
	Action   fwtypes.StringEnum[awstypes.GuardrailWordPolicyAction] `tfsdk:"action"`
	Detected types.Bool                                             `tfsdk:"detected"`
	Match    types.String                                           `tfsdk:"match"`
}

type guardrailManagedWordModel struct {
	Action   fwtypes.StringEnum[awstypes.GuardrailWordPolicyAction] `tfsdk:"action"`
	Detected types.Bool                                             `tfsdk:"detected"`
	Match    types.String                                           `tfsdk:"match"`
	Type     fwtypes.StringEnum[awstypes.GuardrailManagedWordType]  `tfsdk:"type"`
}

type guardrailAutomatedReasoningPolicyAssessmentModel struct {
	Findings fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningFindingModel] `tfsdk:"findings"`
}

type guardrailAutomatedReasoningFindingModel struct {
	Impossible           fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningImpossibleFindingModel]           `tfsdk:"impossible"`
	Invalid              fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningInvalidFindingModel]              `tfsdk:"invalid"`
	NoTranslations       fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningNoTranslationsFindingModel]       `tfsdk:"no_translations"`
	Satisfiable          fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningSatisfiableFindingModel]          `tfsdk:"satisfiable"`
	TooComplex           fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTooComplexFindingModel]           `tfsdk:"too_complex"`
	TranslationAmbiguous fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationAmbiguousFindingModel] `tfsdk:"translation_ambiguous"`
	Valid                fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningValidFindingModel]                `tfsdk:"valid"`
}

var _ flex.Flattener = &guardrailAutomatedReasoningFindingModel{}

func (m *guardrailAutomatedReasoningFindingModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.GuardrailAutomatedReasoningFindingMemberImpossible:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.Impossible))
	case awstypes.GuardrailAutomatedReasoningFindingMemberInvalid:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.Invalid))
	case awstypes.GuardrailAutomatedReasoningFindingMemberNoTranslations:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.NoTranslations))
	case awstypes.GuardrailAutomatedReasoningFindingMemberSatisfiable:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.Satisfiable))
	case awstypes.GuardrailAutomatedReasoningFindingMemberTooComplex:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.TooComplex))
	case awstypes.GuardrailAutomatedReasoningFindingMemberTranslationAmbiguous:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.TranslationAmbiguous))
	case awstypes.GuardrailAutomatedReasoningFindingMemberValid:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.Value, &m.Valid))
	default:
		smerr.AddError(ctx, &diags, fmt.Errorf("unsupported automated reasoning finding type %T", v))
	}

	return diags
}

type guardrailAutomatedReasoningImpossibleFindingModel struct {
	ContradictingRules fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningRuleModel]         `tfsdk:"contradicting_rules"`
	LogicWarning       fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningLogicWarningModel] `tfsdk:"logic_warning"`
	Translation        fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationModel]  `tfsdk:"translation"`
}

type guardrailAutomatedReasoningInvalidFindingModel struct {
	ContradictingRules fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningRuleModel]         `tfsdk:"contradicting_rules"`
	LogicWarning       fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningLogicWarningModel] `tfsdk:"logic_warning"`
	Translation        fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationModel]  `tfsdk:"translation"`
}

type guardrailAutomatedReasoningNoTranslationsFindingModel struct{}

type guardrailAutomatedReasoningSatisfiableFindingModel struct {
	ClaimsFalseScenario fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningScenarioModel]     `tfsdk:"claims_false_scenario"`
	ClaimsTrueScenario  fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningScenarioModel]     `tfsdk:"claims_true_scenario"`
	LogicWarning        fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningLogicWarningModel] `tfsdk:"logic_warning"`
	Translation         fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationModel]  `tfsdk:"translation"`
}

type guardrailAutomatedReasoningTooComplexFindingModel struct{}

type guardrailAutomatedReasoningTranslationAmbiguousFindingModel struct {
	DifferenceScenarios fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningScenarioModel]          `tfsdk:"difference_scenarios"`
	Options             fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationOptionModel] `tfsdk:"options"`
}

type guardrailAutomatedReasoningValidFindingModel struct {
	ClaimsTrueScenario fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningScenarioModel]     `tfsdk:"claims_true_scenario"`
	LogicWarning       fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningLogicWarningModel] `tfsdk:"logic_warning"`
	SupportingRules    fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningRuleModel]         `tfsdk:"supporting_rules"`
	Translation        fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationModel]  `tfsdk:"translation"`
}

type guardrailAutomatedReasoningRuleModel struct {
	Identifier       types.String `tfsdk:"identifier"`
	PolicyVersionARN types.String `tfsdk:"policy_version_arn"`
}

type guardrailAutomatedReasoningLogicWarningModel struct {
	Claims   fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningStatementModel] `tfsdk:"claims"`
	Premises fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningStatementModel] `tfsdk:"premises"`
	Type     fwtypes.StringEnum[awstypes.GuardrailAutomatedReasoningLogicWarningType]   `tfsdk:"type"`
}

type guardrailAutomatedReasoningTranslationModel struct {
	Claims               fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningStatementModel]          `tfsdk:"claims"`
	Confidence           types.Float64                                                                       `tfsdk:"confidence"`
	Premises             fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningStatementModel]          `tfsdk:"premises"`
	UntranslatedClaims   fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningInputTextReferenceModel] `tfsdk:"untranslated_claims"`
	UntranslatedPremises fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningInputTextReferenceModel] `tfsdk:"untranslated_premises"`
}

type guardrailAutomatedReasoningStatementModel struct {
	Logic           types.String `tfsdk:"logic"`
	NaturalLanguage types.String `tfsdk:"natural_language"`
}

type guardrailAutomatedReasoningInputTextReferenceModel struct {
	Text types.String `tfsdk:"text"`
}

type guardrailAutomatedReasoningScenarioModel struct {
	Statements fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningStatementModel] `tfsdk:"statements"`
}

type guardrailAutomatedReasoningTranslationOptionModel struct {
	Translations fwtypes.ListNestedObjectValueOf[guardrailAutomatedReasoningTranslationModel] `tfsdk:"translations"`
}
