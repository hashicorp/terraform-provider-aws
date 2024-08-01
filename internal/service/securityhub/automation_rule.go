// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Automation Rule")
// @Tags(identifierAttribute="arn")
func newAutomationRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &automationRuleResource{}, nil
}

type automationRuleResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *automationRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_securityhub_automation_rule"
}

func (r *automationRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"is_terminal": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"rule_name": schema.StringAttribute{
				Required: true,
			},
			"rule_order": schema.Int64Attribute{
				Required: true,
			},
			"rule_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleStatus](),
				Computed:   true,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrActions: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[automationRulesActionModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AutomationRulesActionType](),
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"finding_fields_update": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[automationRulesFindingFieldsUpdateModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"confidence": schema.Int64Attribute{
										Optional: true,
									},
									"criticality": schema.Int64Attribute{
										Optional: true,
									},
									"types": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									"user_defined_fields": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									"verification_state": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.VerificationState](),
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"note": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[noteUpdateModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"text": schema.StringAttribute{
													Required: true,
												},
												"updated_by": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"related_findings": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[relatedFindingModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrID: schema.StringAttribute{
													Required: true,
												},
												"product_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
											},
										},
									},
									"severity": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[severityUpdateModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"label": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.SeverityLabel](),
													Optional:   true,
													Computed:   true,
												},
												"product": schema.Float64Attribute{
													Optional: true,
												},
											},
										},
									},
									"workflow": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[workflowUpdateModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrStatus: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.WorkflowStatus](),
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"criteria": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[automationRulesFindingFiltersModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrAWSAccountID:               stringFilterSchemaFramework(ctx),
						"aws_account_name":                   stringFilterSchemaFramework(ctx),
						"company_name":                       stringFilterSchemaFramework(ctx),
						"compliance_associated_standards_id": stringFilterSchemaFramework(ctx),
						"compliance_security_control_id":     stringFilterSchemaFramework(ctx),
						"compliance_status":                  stringFilterSchemaFramework(ctx),
						"confidence":                         numberFilterSchemaFramework(ctx),
						names.AttrCreatedAt:                  dateFilterSchemaFramework(ctx),
						"criticality":                        numberFilterSchemaFramework(ctx),
						names.AttrDescription:                stringFilterSchemaFramework(ctx),
						"first_observed_at":                  dateFilterSchemaFramework(ctx),
						"generator_id":                       stringFilterSchemaFramework(ctx),
						names.AttrID:                         stringFilterSchemaFramework(ctx),
						"last_observed_at":                   dateFilterSchemaFramework(ctx),
						"note_text":                          stringFilterSchemaFramework(ctx),
						"note_updated_at":                    dateFilterSchemaFramework(ctx),
						"note_updated_by":                    stringFilterSchemaFramework(ctx),
						"product_arn":                        stringFilterSchemaFramework(ctx),
						"product_name":                       stringFilterSchemaFramework(ctx),
						"record_state":                       stringFilterSchemaFramework(ctx),
						"related_findings_id":                stringFilterSchemaFramework(ctx),
						"related_findings_product_arn":       stringFilterSchemaFramework(ctx),
						"resource_application_arn":           stringFilterSchemaFramework(ctx),
						"resource_application_name":          stringFilterSchemaFramework(ctx),
						"resource_details_other":             mapFilterSchemaFramework(ctx),
						names.AttrResourceID:                 stringFilterSchemaFramework(ctx),
						"resource_partition":                 stringFilterSchemaFramework(ctx),
						"resource_region":                    stringFilterSchemaFramework(ctx),
						names.AttrResourceTags:               mapFilterSchemaFramework(ctx),
						names.AttrResourceType:               stringFilterSchemaFramework(ctx),
						"severity_label":                     stringFilterSchemaFramework(ctx),
						"source_url":                         stringFilterSchemaFramework(ctx),
						"title":                              stringFilterSchemaFramework(ctx),
						names.AttrType:                       stringFilterSchemaFramework(ctx),
						"updated_at":                         dateFilterSchemaFramework(ctx),
						"user_defined_fields":                mapFilterSchemaFramework(ctx),
						"verification_state":                 stringFilterSchemaFramework(ctx),
						"workflow_status":                    stringFilterSchemaFramework(ctx),
					},
				},
			},
		},
	}
}

func dateFilterSchemaFramework(ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[dateFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"end": schema.StringAttribute{
					CustomType: timetypes.RFC3339Type{},
					Optional:   true,
				},
				"start": schema.StringAttribute{
					CustomType: timetypes.RFC3339Type{},
					Optional:   true,
				},
			},
			Blocks: map[string]schema.Block{
				"date_range": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dateRangeModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrUnit: schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.DateRangeUnit](),
								Required:   true,
							},
							names.AttrValue: schema.Int64Attribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func mapFilterSchemaFramework(ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[mapFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"comparison": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.MapFilterComparison](),
					Required:   true,
				},
				names.AttrKey: schema.StringAttribute{
					Required: true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func numberFilterSchemaFramework(ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[numberFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"eq": schema.Float64Attribute{
					Optional: true,
				},
				"gt": schema.Float64Attribute{
					Optional: true,
				},
				"gte": schema.Float64Attribute{
					Optional: true,
				},
				"lt": schema.Float64Attribute{
					Optional: true,
				},
				"lte": schema.Float64Attribute{
					Optional: true,
				},
			},
		},
	}
}

func stringFilterSchemaFramework(ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[stringFilterModel](ctx),
		Validators: []validator.Set{
			setvalidator.SizeAtMost(20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"comparison": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.StringFilterComparison](),
					Required:   true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func (r *automationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data automationRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	input := &securityhub.CreateAutomationRuleInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAutomationRule(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Security Hub Automation Rule (%s)", aws.ToString(input.RuleName)), err.Error())

		return
	}

	// Set values for unknowns.
	ruleARN := aws.ToString(output.RuleArn)
	data.RuleARN = types.StringValue(ruleARN)
	data.setID()

	automationRule, err := findAutomationRuleByARN(ctx, conn, ruleARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub Automation Rule (%s)", ruleARN), err.Error())

		return
	}

	data.RuleStatus = fwtypes.StringEnumValue(automationRule.RuleStatus)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *automationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data automationRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	ruleARN := data.ID.ValueString()
	automationRule, err := findAutomationRuleByARN(ctx, conn, ruleARN)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub Automation Rule (%s)", ruleARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, automationRule, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *automationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new automationRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	if !new.Actions.Equal(old.Actions) ||
		!new.Criteria.Equal(old.Criteria) ||
		!new.Description.Equal(old.Description) ||
		!new.IsTerminal.Equal(old.IsTerminal) ||
		!new.RuleName.Equal(old.RuleName) ||
		!new.RuleOrder.Equal(old.RuleOrder) ||
		!new.RuleStatus.Equal(old.RuleStatus) {
		item := awstypes.UpdateAutomationRulesRequestItem{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &item)...)
		if response.Diagnostics.HasError() {
			return
		}

		input := &securityhub.BatchUpdateAutomationRulesInput{
			UpdateAutomationRulesRequestItems: []awstypes.UpdateAutomationRulesRequestItem{item},
		}

		_, err := conn.BatchUpdateAutomationRules(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Hub Automation Rule (%s)", aws.ToString(item.RuleArn)), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *automationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data automationRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	ruleARN := data.ID.ValueString()
	_, err := conn.BatchDeleteAutomationRules(ctx, &securityhub.BatchDeleteAutomationRulesInput{
		AutomationRulesArns: []string{ruleARN},
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Hub Automation Rule (%s)", ruleARN), err.Error())

		return
	}
}

func (r *automationRuleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findAutomationRuleByARN(ctx context.Context, conn *securityhub.Client, arn string) (*awstypes.AutomationRulesConfig, error) {
	input := &securityhub.BatchGetAutomationRulesInput{
		AutomationRulesArns: []string{arn},
	}

	return findAutomationRule(ctx, conn, input)
}

func findAutomationRule(ctx context.Context, conn *securityhub.Client, input *securityhub.BatchGetAutomationRulesInput) (*awstypes.AutomationRulesConfig, error) {
	output, err := findAutomationRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAutomationRules(ctx context.Context, conn *securityhub.Client, input *securityhub.BatchGetAutomationRulesInput) ([]awstypes.AutomationRulesConfig, error) {
	output, err := conn.BatchGetAutomationRules(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Rules, nil
}

type automationRuleResourceModel struct {
	Actions     fwtypes.SetNestedObjectValueOf[automationRulesActionModel]          `tfsdk:"actions"`
	Criteria    fwtypes.ListNestedObjectValueOf[automationRulesFindingFiltersModel] `tfsdk:"criteria"`
	Description types.String                                                        `tfsdk:"description"`
	ID          types.String                                                        `tfsdk:"id"`
	IsTerminal  types.Bool                                                          `tfsdk:"is_terminal"`
	RuleARN     types.String                                                        `tfsdk:"arn"`
	RuleName    types.String                                                        `tfsdk:"rule_name"`
	RuleOrder   types.Int64                                                         `tfsdk:"rule_order"`
	RuleStatus  fwtypes.StringEnum[awstypes.RuleStatus]                             `tfsdk:"rule_status"`
	Tags        types.Map                                                           `tfsdk:"tags"`
	TagsAll     types.Map                                                           `tfsdk:"tags_all"`
}

func (data *automationRuleResourceModel) InitFromID() error {
	data.RuleARN = data.ID

	return nil
}

func (data *automationRuleResourceModel) setID() {
	data.ID = data.RuleARN
}

type automationRulesActionModel struct {
	FindingFieldsUpdate fwtypes.ListNestedObjectValueOf[automationRulesFindingFieldsUpdateModel] `tfsdk:"finding_fields_update"`
	Type                fwtypes.StringEnum[awstypes.AutomationRulesActionType]                   `tfsdk:"type"`
}

type automationRulesFindingFieldsUpdateModel struct {
	Confidence        types.Int64                                          `tfsdk:"confidence"`
	Criticality       types.Int64                                          `tfsdk:"criticality"`
	Note              fwtypes.ListNestedObjectValueOf[noteUpdateModel]     `tfsdk:"note"`
	RelatedFindings   fwtypes.SetNestedObjectValueOf[relatedFindingModel]  `tfsdk:"related_findings"`
	Severity          fwtypes.ListNestedObjectValueOf[severityUpdateModel] `tfsdk:"severity"`
	Types             fwtypes.ListValueOf[types.String]                    `tfsdk:"types"`
	UserDefinedFields fwtypes.MapValueOf[types.String]                     `tfsdk:"user_defined_fields"`
	VerificationState fwtypes.StringEnum[awstypes.VerificationState]       `tfsdk:"verification_state"`
	Workflow          fwtypes.ListNestedObjectValueOf[workflowUpdateModel] `tfsdk:"workflow"`
}

type noteUpdateModel struct {
	Text      types.String `tfsdk:"text"`
	UpdatedBy types.String `tfsdk:"updated_by"`
}

type relatedFindingModel struct {
	ID         types.String `tfsdk:"id"`
	ProductARN fwtypes.ARN  `tfsdk:"product_arn"`
}

type severityUpdateModel struct {
	Label   fwtypes.StringEnum[awstypes.SeverityLabel] `tfsdk:"label"`
	Product types.Float64                              `tfsdk:"product"`
}

type workflowUpdateModel struct {
	Status fwtypes.StringEnum[awstypes.WorkflowStatus] `tfsdk:"status"`
}

type automationRulesFindingFiltersModel struct {
	AWSAccountID                    fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"aws_account_id"`
	AWSAccountName                  fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"aws_account_name"`
	CompanyName                     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"company_name"`
	ComplianceAssociatedStandardsID fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"compliance_associated_standards_id"`
	ComplianceSecurityControlID     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"compliance_security_control_id"`
	ComplianceStatus                fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"compliance_status"`
	Confidence                      fwtypes.SetNestedObjectValueOf[numberFilterModel] `tfsdk:"confidence"`
	CreatedAt                       fwtypes.SetNestedObjectValueOf[dateFilterModel]   `tfsdk:"created_at"`
	Criticality                     fwtypes.SetNestedObjectValueOf[numberFilterModel] `tfsdk:"criticality"`
	Description                     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"description"`
	FirstObservedAt                 fwtypes.SetNestedObjectValueOf[dateFilterModel]   `tfsdk:"first_observed_at"`
	GeneratorID                     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"generator_id"`
	ID                              fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"id"`
	LastObservedAt                  fwtypes.SetNestedObjectValueOf[dateFilterModel]   `tfsdk:"last_observed_at"`
	NoteText                        fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"note_text"`
	NoteUpdatedAt                   fwtypes.SetNestedObjectValueOf[dateFilterModel]   `tfsdk:"note_updated_at"`
	NoteUpdatedBy                   fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"note_updated_by"`
	ProductARN                      fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"product_arn"`
	ProductName                     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"product_name"`
	RecordState                     fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"record_state"`
	RelatedFindingsID               fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"related_findings_id"`
	RelatedFindingsProductARN       fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"related_findings_product_arn"`
	ResourceApplicationARN          fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_application_arn"`
	ResourceApplicationName         fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_application_name"`
	ResourceDetailsOther            fwtypes.SetNestedObjectValueOf[mapFilterModel]    `tfsdk:"resource_details_other"`
	ResourceID                      fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_id"`
	ResourcePartition               fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_partition"`
	ResourceRegion                  fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_region"`
	ResourceTags                    fwtypes.SetNestedObjectValueOf[mapFilterModel]    `tfsdk:"resource_tags"`
	ResourceType                    fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"resource_type"`
	SeverityLabel                   fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"severity_label"`
	SourceUrl                       fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"source_url"`
	Title                           fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"title"`
	Type                            fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"type"`
	UpdatedAt                       fwtypes.SetNestedObjectValueOf[dateFilterModel]   `tfsdk:"updated_at"`
	UserDefinedFields               fwtypes.SetNestedObjectValueOf[mapFilterModel]    `tfsdk:"user_defined_fields"`
	VerificationState               fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"verification_state"`
	WorkflowStatus                  fwtypes.SetNestedObjectValueOf[stringFilterModel] `tfsdk:"workflow_status"`
}

type dateFilterModel struct {
	DateRange fwtypes.ListNestedObjectValueOf[dateRangeModel] `tfsdk:"date_range"`
	End       timetypes.RFC3339                               `tfsdk:"end"`
	Start     timetypes.RFC3339                               `tfsdk:"start"`
}

type dateRangeModel struct {
	Unit  fwtypes.StringEnum[awstypes.DateRangeUnit] `tfsdk:"unit"`
	Value types.Int64                                `tfsdk:"value"`
}

type mapFilterModel struct {
	Comparison fwtypes.StringEnum[awstypes.MapFilterComparison] `tfsdk:"comparison"`
	Key        types.String                                     `tfsdk:"key"`
	Value      types.String                                     `tfsdk:"value"`
}

type numberFilterModel struct {
	EQ  types.Float64 `tfsdk:"eq"`
	GT  types.Float64 `tfsdk:"gt"`
	GTE types.Float64 `tfsdk:"gte"`
	LT  types.Float64 `tfsdk:"lt"`
	LTE types.Float64 `tfsdk:"lte"`
}

type stringFilterModel struct {
	Comparison fwtypes.StringEnum[awstypes.StringFilterComparison] `tfsdk:"comparison"`
	Value      types.String                                        `tfsdk:"value"`
}
