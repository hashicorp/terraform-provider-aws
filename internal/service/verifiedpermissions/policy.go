// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/cedar-policy/cedar-go"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	interflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(aws_verifiedpermissions_policy, name="Policy")
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyResource{}

	return r, nil
}

const (
	ResNamePolicy = "Policy"
)

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithImportByID
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"policy_id":  framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyDefinition](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"static": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[staticPolicyDefinition](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("template_linked"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDescription: schema.StringAttribute{
										Optional: true,
									},
									"statement": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplaceIf(
												statementReplaceIf, "Replace cedar statement diff", "Replace cedar statement diff",
											),
										},
									},
								},
							},
						},
						"template_linked": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[templateLinkedPolicyDefinition](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("static"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"policy_template_id": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrPrincipal: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[templateLinkedPrincipal](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"entity_id": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"entity_type": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
									"resource": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[templateLinkedResource](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"entity_id": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"entity_type": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
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
		},
	}
}

func statementReplaceIf(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.State.Raw.IsNull() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	// Parse the plan policy
	planPolicies, err := cedar.NewPolicyListFromBytes("plan.cedar", []byte(req.PlanValue.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse plan policy", err.Error())
		return
	}

	// Parse the state policy
	statePolicies, err := cedar.NewPolicyListFromBytes("state.cedar", []byte(req.StateValue.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse state policy", err.Error())
		return
	}

	var policyPrincipal, policyResource, policyEffect bool
	if len(planPolicies) > 0 && len(statePolicies) > 0 {
		planPolicyAST := planPolicies[0].AST()
		statePolicyAST := statePolicies[0].AST()

		policyEffect = planPolicyAST.Effect != statePolicyAST.Effect
		policyPrincipal = planPolicyAST.Principal != statePolicyAST.Principal
		policyResource = planPolicyAST.Resource != statePolicyAST.Resource
	}

	resp.RequiresReplace = policyEffect || policyPrincipal || policyResource
}

const (
	ResourcePolicyIDPartsCount = 2
)

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var plan policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &verifiedpermissions.CreatePolicyInput{}

	in.ClientToken = aws.String(id.UniqueId())
	in.PolicyStoreId = plan.PolicyStoreID.ValueStringPointer()

	def, diags := plan.Definition.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if !def.Static.IsNull() {
		static, diags := def.Static.ToPtr(ctx)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		in.Definition = &awstypes.PolicyDefinitionMemberStatic{
			Value: awstypes.StaticPolicyDefinition{
				Statement:   fwflex.StringFromFramework(ctx, static.Statement),
				Description: fwflex.StringFromFramework(ctx, static.Description),
			},
		}
	}

	if !def.TemplateLinked.IsNull() {
		templateLinked, diags := def.TemplateLinked.ToPtr(ctx)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		value := awstypes.TemplateLinkedPolicyDefinition{
			PolicyTemplateId: templateLinked.PolicyTemplateID.ValueStringPointer(),
		}

		if !templateLinked.Principal.IsNull() {
			principal, diags := templateLinked.Principal.ToPtr(ctx)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}

			value.Principal = &awstypes.EntityIdentifier{
				EntityId:   fwflex.StringFromFramework(ctx, principal.EntityID),
				EntityType: fwflex.StringFromFramework(ctx, principal.EntityType),
			}
		}

		if !templateLinked.Resource.IsNull() {
			res, diags := templateLinked.Resource.ToPtr(ctx)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}

			value.Resource = &awstypes.EntityIdentifier{
				EntityId:   fwflex.StringFromFramework(ctx, res.EntityID),
				EntityType: fwflex.StringFromFramework(ctx, res.EntityType),
			}
		}

		in.Definition = &awstypes.PolicyDefinitionMemberTemplateLinked{
			Value: value,
		}
	}

	out, err := conn.CreatePolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicy, plan.PolicyStoreID.String(), err),
			err.Error(),
		)
		return
	}

	idParts := []string{
		aws.ToString(out.PolicyId),
		aws.ToString(out.PolicyStoreId),
	}

	rID, err := interflex.FlattenResourceId(idParts, ResourcePolicyIDPartsCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicy, plan.PolicyStoreID.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, rID)
	plan.CreatedDate = timetypes.NewRFC3339TimePointerValue(out.CreatedDate)
	plan.PolicyID = fwflex.StringToFramework(ctx, out.PolicyId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var state policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rID, err := interflex.ExpandResourceId(state.ID.ValueString(), ResourcePolicyIDPartsCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionSetting, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findPolicyByID(ctx, conn, rID[0], rID[1])
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionSetting, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.PolicyID = fwflex.StringToFramework(ctx, out.PolicyId)
	state.PolicyStoreID = fwflex.StringToFramework(ctx, out.PolicyStoreId)
	state.CreatedDate = timetypes.NewRFC3339TimePointerValue(out.CreatedDate)

	if val, ok := out.Definition.(*awstypes.PolicyDefinitionDetailMemberStatic); ok && val != nil {
		static := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &staticPolicyDefinition{
			Statement:   fwflex.StringToFramework(ctx, val.Value.Statement),
			Description: fwflex.StringToFramework(ctx, val.Value.Description),
		})

		state.Definition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &policyDefinition{
			Static:         static,
			TemplateLinked: fwtypes.NewListNestedObjectValueOfNull[templateLinkedPolicyDefinition](ctx),
		})
	}

	if val, ok := out.Definition.(*awstypes.PolicyDefinitionDetailMemberTemplateLinked); ok && val != nil {
		tpl := templateLinkedPolicyDefinition{
			PolicyTemplateID: fwflex.StringToFramework(ctx, val.Value.PolicyTemplateId),
		}

		if val.Value.Principal != nil {
			tpl.Principal = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &templateLinkedPrincipal{
				EntityID:   fwflex.StringToFramework(ctx, val.Value.Principal.EntityId),
				EntityType: fwflex.StringToFramework(ctx, val.Value.Principal.EntityType),
			})
		} else {
			tpl.Principal = fwtypes.NewListNestedObjectValueOfNull[templateLinkedPrincipal](ctx)
		}

		if val.Value.Resource != nil {
			tpl.Resource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &templateLinkedResource{
				EntityID:   fwflex.StringToFramework(ctx, val.Value.Resource.EntityId),
				EntityType: fwflex.StringToFramework(ctx, val.Value.Resource.EntityType),
			})
		} else {
			tpl.Resource = fwtypes.NewListNestedObjectValueOfNull[templateLinkedResource](ctx)
		}

		templateLinked := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tpl)

		state.Definition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &policyDefinition{
			Static:         fwtypes.NewListNestedObjectValueOfNull[staticPolicyDefinition](ctx),
			TemplateLinked: templateLinked,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var plan, state policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Definition.Equal(state.Definition) {
		in := &verifiedpermissions.UpdatePolicyInput{}
		in.PolicyId = fwflex.StringFromFramework(ctx, state.PolicyID)
		in.PolicyStoreId = fwflex.StringFromFramework(ctx, state.PolicyStoreID)

		defPlan, diagsPlan := plan.Definition.ToPtr(ctx)
		resp.Diagnostics.Append(diagsPlan...)
		if resp.Diagnostics.HasError() {
			return
		}

		defState, diagsState := state.Definition.ToPtr(ctx)
		resp.Diagnostics.Append(diagsState...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !defPlan.Static.Equal(defState.Static) {
			static, diags := defPlan.Static.ToPtr(ctx)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}

			in.Definition = &awstypes.UpdatePolicyDefinitionMemberStatic{
				Value: awstypes.UpdateStaticPolicyDefinition{
					Statement:   fwflex.StringFromFramework(ctx, static.Statement),
					Description: fwflex.StringFromFramework(ctx, static.Description),
				},
			}
		}

		_, err := conn.UpdatePolicy(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var state policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &verifiedpermissions.DeletePolicyInput{
		PolicyId:      state.PolicyID.ValueStringPointer(),
		PolicyStoreId: state.PolicyStoreID.ValueStringPointer(),
	}

	_, err := conn.DeletePolicy(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *policyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var plan, state policyResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Definition.Equal(state.Definition) {
			defPlan, diagsPlan := plan.Definition.ToPtr(ctx)
			resp.Diagnostics.Append(diagsPlan...)
			if resp.Diagnostics.HasError() {
				return
			}
			defState, diagsState := state.Definition.ToPtr(ctx)
			resp.Diagnostics.Append(diagsState...)
			if resp.Diagnostics.HasError() {
				return
			}

			if !defState.Static.IsNull() && defPlan.Static.IsNull() {
				resp.RequiresReplace = []path.Path{path.Root("definition").AtListIndex(0).AtName("static")}
			}

			if !defState.TemplateLinked.IsNull() && defPlan.TemplateLinked.IsNull() {
				resp.RequiresReplace = []path.Path{path.Root("definition").AtListIndex(0).AtName("template_linked")}
			}
		}
	}
}

func findPolicyByID(ctx context.Context, conn *verifiedpermissions.Client, id, policyStoreId string) (*verifiedpermissions.GetPolicyOutput, error) {
	in := &verifiedpermissions.GetPolicyInput{
		PolicyId:      aws.String(id),
		PolicyStoreId: aws.String(policyStoreId),
	}

	out, err := conn.GetPolicy(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.PolicyId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type policyResourceModel struct {
	framework.WithRegionModel
	CreatedDate   timetypes.RFC3339                                 `tfsdk:"created_date"`
	Definition    fwtypes.ListNestedObjectValueOf[policyDefinition] `tfsdk:"definition"`
	ID            types.String                                      `tfsdk:"id"`
	PolicyID      types.String                                      `tfsdk:"policy_id"`
	PolicyStoreID types.String                                      `tfsdk:"policy_store_id"`
}

type policyDefinition struct {
	Static         fwtypes.ListNestedObjectValueOf[staticPolicyDefinition]         `tfsdk:"static"`
	TemplateLinked fwtypes.ListNestedObjectValueOf[templateLinkedPolicyDefinition] `tfsdk:"template_linked"`
}

type staticPolicyDefinition struct {
	Statement   types.String `tfsdk:"statement"`
	Description types.String `tfsdk:"description"`
}

type templateLinkedPolicyDefinition struct {
	PolicyTemplateID types.String                                             `tfsdk:"policy_template_id"`
	Principal        fwtypes.ListNestedObjectValueOf[templateLinkedPrincipal] `tfsdk:"principal"`
	Resource         fwtypes.ListNestedObjectValueOf[templateLinkedResource]  `tfsdk:"resource"`
}

type templateLinkedPrincipal struct {
	EntityID   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
}

type templateLinkedResource struct {
	EntityID   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
}
