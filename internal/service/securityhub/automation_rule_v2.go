// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_automation_rule_v2", name="Automation Rule V2")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub;securityhub;securityhub.GetAutomationRuleV2Output")
// @Testing(serialize=true)
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
func newAutomationRuleV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &automationRuleV2Resource{}, nil
}

type automationRuleV2Resource struct {
	framework.ResourceWithModel[automationRuleV2ResourceModel]
	framework.WithImportByIdentity
}

func (r *automationRuleV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Required:    true,
				Description: "A description of the automation rule.",
			},
			"rule_id": framework.IDAttribute(),
			"rule_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the automation rule.",
			},
			"rule_order": schema.Float64Attribute{
				Required:    true,
				Description: "The priority of the rule (lower values = higher priority).",
				Validators: []validator.Float64{
					float64validator.Between(1.0, 1000.0),
				},
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"rule_status": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.RuleStatusV2](),
				Optional:    true,
				Computed:    true,
				Description: "The status of the rule: ENABLED or DISABLED.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[automationRulesActionV2Model](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				Description: "Actions to take when the rule matches. Maximum of 1 action.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType:  fwtypes.StringEnumType[awstypes.AutomationRulesActionTypeV2](),
							Required:    true,
							Description: "The action type: FINDING_FIELDS_UPDATE or EXTERNAL_INTEGRATION.",
						},
					},
					Blocks: map[string]schema.Block{
						"external_integration_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[externalIntegrationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							Description: "Settings for external integration actions.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"connector_arn": schema.StringAttribute{
										CustomType:  fwtypes.ARNType,
										Required:    true,
										Description: "The ARN of the connector.",
									},
								},
							},
						},
						"finding_fields_update": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[automationRulesFindingFieldsUpdateV2Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							Description: "Settings for updating finding fields.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrComment: schema.StringAttribute{
										Optional:    true,
										Description: "A comment for the finding.",
									},
									"severity_id": schema.Int32Attribute{
										Optional:    true,
										Description: "The severity ID to assign.",
									},
									"status_id": schema.Int32Attribute{
										Optional:    true,
										Description: "The status ID to assign.",
									},
								},
							},
						},
					},
				},
			},
			"criteria": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[criteriaModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				Description: "Filtering type and configuration of the automation rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ocsf_finding_criteria_json": schema.StringAttribute{
							CustomType:  jsontypes.NormalizedType{},
							Required:    true,
							Description: "JSON-encoded OCSF finding criteria for the rule.",
						},
					},
				},
			},
		},
	}
}

func (r *automationRuleV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data automationRuleV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.RuleName)
	var input securityhub.CreateAutomationRuleV2Input
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	criteria, diags := data.Criteria.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if criteria != nil {
		var ocsfFindingFilters awstypes.OcsfFindingFilters
		if err := tfjson.DecodeFromString(fwflex.StringValueFromFramework(ctx, criteria.OCSFFindingCriteriaJSON), &ocsfFindingFilters); err != nil {
			response.Diagnostics.AddError("invalid ocsf_finding_criteria_json", err.Error())
			return
		}
		input.Criteria = &awstypes.CriteriaMemberOcsfFindingCriteria{Value: ocsfFindingFilters}
	}

	output, err := conn.CreateAutomationRuleV2(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Security Hub V2 Automation Rule (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	data.RuleARN = fwflex.StringToFramework(ctx, output.RuleArn)
	data.RuleID = fwflex.StringToFramework(ctx, output.RuleId)
	if data.RuleStatus.IsUnknown() {
		data.RuleStatus = fwtypes.StringEnumValue(awstypes.RuleStatusV2Enabled)
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *automationRuleV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data automationRuleV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.RuleARN)
	output, err := findAutomationRuleV2ByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Automation Rule (%s)", arn), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *automationRuleV2Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new automationRuleV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	diff, diags := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, new.RuleARN)
		var input securityhub.UpdateAutomationRuleV2Input
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.Identifier = aws.String(arn)

		criteria, diags := new.Criteria.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if criteria != nil {
			var ocsfFindingFilters awstypes.OcsfFindingFilters
			if err := tfjson.DecodeFromString(fwflex.StringValueFromFramework(ctx, criteria.OCSFFindingCriteriaJSON), &ocsfFindingFilters); err != nil {
				response.Diagnostics.AddError("invalid ocsf_finding_criteria_json", err.Error())
				return
			}
			input.Criteria = &awstypes.CriteriaMemberOcsfFindingCriteria{Value: ocsfFindingFilters}
		}

		_, err := conn.UpdateAutomationRuleV2(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Hub V2 Automation Rule (%s)", arn), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *automationRuleV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data automationRuleV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.RuleARN)
	input := securityhub.DeleteAutomationRuleV2Input{
		Identifier: aws.String(arn),
	}
	_, err := conn.DeleteAutomationRuleV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Hub V2 Automation Rule (%s)", arn), err.Error())
	}
}

func (r *automationRuleV2Resource) flatten(ctx context.Context, automationRuleV2 *securityhub.GetAutomationRuleV2Output, data *automationRuleV2ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, automationRuleV2, data)...)

	if v := automationRuleV2.Criteria; v != nil {
		switch t := v.(type) {
		case *awstypes.CriteriaMemberOcsfFindingCriteria:
			v, err := tfjson.EncodeToBytes(t.Value)
			if err != nil {
				diags.AddError("invalid ocsf_finding_criteria_json", err.Error())
				return diags
			}
			criteria, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &criteriaModel{
				OCSFFindingCriteriaJSON: jsontypes.NewNormalizedValue(string(tfjson.RemoveEmptyStringFields(tfjson.RemoveEmptyFields(v)))),
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			data.Criteria = criteria
		default:
		}
	}

	return diags
}

func findAutomationRuleV2ByARN(ctx context.Context, conn *securityhub.Client, arn string) (*securityhub.GetAutomationRuleV2Output, error) {
	input := securityhub.GetAutomationRuleV2Input{
		Identifier: aws.String(arn),
	}

	return findAutomationRuleV2(ctx, conn, &input)
}

func findAutomationRuleV2(ctx context.Context, conn *securityhub.Client, input *securityhub.GetAutomationRuleV2Input) (*securityhub.GetAutomationRuleV2Output, error) {
	output, err := conn.GetAutomationRuleV2(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Security Hub V2 is not enabled") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type automationRuleV2ResourceModel struct {
	framework.WithRegionModel
	Actions     fwtypes.ListNestedObjectValueOf[automationRulesActionV2Model] `tfsdk:"action"`
	Criteria    fwtypes.ListNestedObjectValueOf[criteriaModel]                `tfsdk:"criteria" autoflex:"-"`
	Description types.String                                                  `tfsdk:"description"`
	RuleARN     types.String                                                  `tfsdk:"arn"`
	RuleID      types.String                                                  `tfsdk:"rule_id"`
	RuleName    types.String                                                  `tfsdk:"rule_name"`
	RuleOrder   types.Float64                                                 `tfsdk:"rule_order"`
	RuleStatus  fwtypes.StringEnum[awstypes.RuleStatusV2]                     `tfsdk:"rule_status"`
	Tags        tftags.Map                                                    `tfsdk:"tags"`
	TagsAll     tftags.Map                                                    `tfsdk:"tags_all"`
}

type automationRulesActionV2Model struct {
	ExternalIntegrationConfiguration fwtypes.ListNestedObjectValueOf[externalIntegrationConfigurationModel]     `tfsdk:"external_integration_configuration"`
	FindingFieldsUpdate              fwtypes.ListNestedObjectValueOf[automationRulesFindingFieldsUpdateV2Model] `tfsdk:"finding_fields_update"`
	Type                             fwtypes.StringEnum[awstypes.AutomationRulesActionTypeV2]                   `tfsdk:"type"`
}

type externalIntegrationConfigurationModel struct {
	ConnectorARN fwtypes.ARN `tfsdk:"connector_arn"`
}

type automationRulesFindingFieldsUpdateV2Model struct {
	Comment    types.String `tfsdk:"comment"`
	SeverityID types.Int32  `tfsdk:"severity_id"`
	StatusID   types.Int32  `tfsdk:"status_id"`
}

type criteriaModel struct {
	OCSFFindingCriteriaJSON jsontypes.Normalized `tfsdk:"ocsf_finding_criteria_json"`
}
