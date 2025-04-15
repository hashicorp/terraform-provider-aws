// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_imagebuilder_lifecycle_policy", name="Lifecycle Policy")
// @Tags(identifierAttribute="id")
func newLifecyclePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &lifecyclePolicyResource{}, nil
}

type lifecyclePolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *lifecyclePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	lifecyclePolicyStatusType := fwtypes.StringEnumType[awstypes.LifecyclePolicyStatus]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"execution_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[-_A-Za-z-0-9][-_A-Za-z0-9 ]{1,126}[-_A-Za-z-0-9]$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrResourceType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LifecyclePolicyResourceType](),
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: lifecyclePolicyStatusType,
				Optional:   true,
				Computed:   true,
				Default:    lifecyclePolicyStatusType.AttributeDefault(awstypes.LifecyclePolicyStatusEnabled),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"policy_detail": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[lifecyclePolicyDetailModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(1, 3),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrAction: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailActionModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.LifecyclePolicyDetailActionType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"include_resources": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailActionIncludeResourcesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amis": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
												"containers": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
												"snapshots": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
								},
							},
						},
						"exclusion_rules": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailExclusionRulesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_map": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"amis": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailExclusionRulesAMIsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"is_public": schema.BoolAttribute{
													Optional: true,
												},
												"regions": schema.ListAttribute{
													CustomType:  fwtypes.ListOfStringType,
													ElementType: types.StringType,
													Optional:    true,
												},
												"shared_accounts": schema.ListAttribute{
													CustomType:  fwtypes.ListOfStringType,
													ElementType: types.StringType,
													Optional:    true,
												},
												"tag_map": schema.MapAttribute{
													CustomType:  fwtypes.MapOfStringType,
													Optional:    true,
													ElementType: types.StringType,
												},
											},
											Blocks: map[string]schema.Block{
												"last_launched": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailExclusionRulesAmisLastLaunchedModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrUnit: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.LifecyclePolicyTimeUnit](),
																Required:   true,
															},
															names.AttrValue: schema.Int64Attribute{
																Required: true,
																Validators: []validator.Int64{
																	int64validator.Between(1, 365),
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
						},
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDetailFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"retain_at_least": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 10),
										},
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.LifecyclePolicyDetailFilterType](),
										Required:   true,
									},
									names.AttrUnit: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.LifecyclePolicyTimeUnit](),
										Optional:   true,
									},
									names.AttrValue: schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 1000),
										},
									},
								},
							},
						},
					},
				},
			},
			"resource_selection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyResourceSelectionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag_map": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						"recipe": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[lifecyclePolicyResourceSelectionRecipeModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtMost(50),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[-_A-Za-z-0-9][-_A-Za-z0-9 ]{1,126}[-_A-Za-z-0-9]$`), ""),
										},
									},
									"semantic_version": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`), ""),
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

func (r *lifecyclePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ImageBuilderClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := &imagebuilder.CreateLifecyclePolicyInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
		return conn.CreateLifecyclePolicy(ctx, input)
	}, errCodeInvalidParameterValueException, "The provided role does not exist or does not have sufficient permissions")

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Image Builder Lifecycle Policy (%s)", name), err.Error())

		return
	}

	output := outputRaw.(*imagebuilder.CreateLifecyclePolicyOutput)

	// Set values for unknowns.
	data.LifecyclePolicyARN = fwflex.StringToFramework(ctx, output.LifecyclePolicyArn)
	data.setID()

	// Read to retrieve computed arguments not part of the create response.
	policy, err := findLifecyclePolicyByARN(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Image Builder Lifecycle Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, policy, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *lifecyclePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ImageBuilderClient(ctx)

	policy, err := findLifecyclePolicyByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Image Builder Lifecycle Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, policy.Tags)

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, policy, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *lifecyclePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ImageBuilderClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.ExecutionRole.Equal(old.ExecutionRole) ||
		!new.PolicyDetails.Equal(old.PolicyDetails) ||
		!new.ResourceSelection.Equal(old.ResourceSelection) ||
		!new.ResourceType.Equal(old.ResourceType) ||
		!new.Status.Equal(old.Status) {
		input := &imagebuilder.UpdateLifecyclePolicyInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
			return conn.UpdateLifecyclePolicy(ctx, input)
		}, errCodeInvalidParameterValueException, "The provided role does not exist or does not have sufficient permissions")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Image Builder Lifecycle Policy (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *lifecyclePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ImageBuilderClient(ctx)

	_, err := conn.DeleteLifecyclePolicy(ctx, &imagebuilder.DeleteLifecyclePolicyInput{
		LifecyclePolicyArn: data.ID.ValueStringPointer(),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Image Builder Lifecycle Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findLifecyclePolicyByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.LifecyclePolicy, error) {
	input := &imagebuilder.GetLifecyclePolicyInput{
		LifecyclePolicyArn: aws.String(arn),
	}

	output, err := conn.GetLifecyclePolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LifecyclePolicy, nil
}

type lifecyclePolicyResourceModel struct {
	Description        types.String                                                           `tfsdk:"description"`
	ExecutionRole      fwtypes.ARN                                                            `tfsdk:"execution_role"`
	ID                 types.String                                                           `tfsdk:"id"`
	LifecyclePolicyARN types.String                                                           `tfsdk:"arn"`
	Name               types.String                                                           `tfsdk:"name"`
	PolicyDetails      fwtypes.SetNestedObjectValueOf[lifecyclePolicyDetailModel]             `tfsdk:"policy_detail"`
	ResourceSelection  fwtypes.ListNestedObjectValueOf[lifecyclePolicyResourceSelectionModel] `tfsdk:"resource_selection"`
	ResourceType       fwtypes.StringEnum[awstypes.LifecyclePolicyResourceType]               `tfsdk:"resource_type"`
	Status             fwtypes.StringEnum[awstypes.LifecyclePolicyStatus]                     `tfsdk:"status"`
	Tags               tftags.Map                                                             `tfsdk:"tags"`
	TagsAll            tftags.Map                                                             `tfsdk:"tags_all"`
}

func (model *lifecyclePolicyResourceModel) InitFromID() error {
	model.LifecyclePolicyARN = model.ID

	return nil
}

func (model *lifecyclePolicyResourceModel) setID() {
	model.ID = model.LifecyclePolicyARN
}

type lifecyclePolicyDetailModel struct {
	Action         fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailActionModel]         `tfsdk:"action"`
	ExclusionRules fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailExclusionRulesModel] `tfsdk:"exclusion_rules"`
	Filter         fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailFilterModel]         `tfsdk:"filter"`
}

type lifecyclePolicyDetailActionModel struct {
	IncludeResources fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailActionIncludeResourcesModel] `tfsdk:"include_resources"`
	Type             fwtypes.StringEnum[awstypes.LifecyclePolicyDetailActionType]                      `tfsdk:"type"`
}

type lifecyclePolicyDetailActionIncludeResourcesModel struct {
	AMIs       types.Bool `tfsdk:"amis"`
	Containers types.Bool `tfsdk:"containers"`
	Snapshots  types.Bool `tfsdk:"snapshots"`
}

type lifecyclePolicyDetailExclusionRulesModel struct {
	AMIs   fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailExclusionRulesAMIsModel] `tfsdk:"amis"`
	TagMap fwtypes.MapValueOf[types.String]                                              `tfsdk:"tag_map"`
}

type lifecyclePolicyDetailExclusionRulesAMIsModel struct {
	IsPublic       types.Bool                                                                                `tfsdk:"is_public"`
	LastLaunched   fwtypes.ListNestedObjectValueOf[lifecyclePolicyDetailExclusionRulesAmisLastLaunchedModel] `tfsdk:"last_launched"`
	Regions        fwtypes.ListValueOf[types.String]                                                         `tfsdk:"regions"`
	SharedAccounts fwtypes.ListValueOf[types.String]                                                         `tfsdk:"shared_accounts"`
	TagMap         fwtypes.MapValueOf[types.String]                                                          `tfsdk:"tag_map"`
}

type lifecyclePolicyDetailExclusionRulesAmisLastLaunchedModel struct {
	Unit  fwtypes.StringEnum[awstypes.LifecyclePolicyTimeUnit] `tfsdk:"unit"`
	Value types.Int64                                          `tfsdk:"value"`
}

type lifecyclePolicyDetailFilterModel struct {
	RetainAtLeast types.Int64                                                  `tfsdk:"retain_at_least"`
	Type          fwtypes.StringEnum[awstypes.LifecyclePolicyDetailFilterType] `tfsdk:"type"`
	Unit          fwtypes.StringEnum[awstypes.LifecyclePolicyTimeUnit]         `tfsdk:"unit"`
	Value         types.Int64                                                  `tfsdk:"value"`
}

type lifecyclePolicyResourceSelectionModel struct {
	Recipes fwtypes.SetNestedObjectValueOf[lifecyclePolicyResourceSelectionRecipeModel] `tfsdk:"recipe"`
	TagMap  fwtypes.MapValueOf[types.String]                                            `tfsdk:"tag_map"`
}

type lifecyclePolicyResourceSelectionRecipeModel struct {
	Name            types.String `tfsdk:"name"`
	SemanticVersion types.String `tfsdk:"semantic_version"`
}
