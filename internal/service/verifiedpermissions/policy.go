// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	interflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(aws_verifiedpermissions_policy, name="Policy")
func newResourcePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicy{}

	return r, nil
}

const (
	ResNamePolicy = "Policy"
)

type resourcePolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourcePolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_verifiedpermissions_policy"
}

func (r *resourcePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"id":        framework.IDAttribute(),
			"policy_id": framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Required: true,
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
									"description": schema.StringAttribute{
										Optional: true,
									},
									"statement": schema.StringAttribute{
										Required: true,
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
									},
								},
								Blocks: map[string]schema.Block{
									"principal": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[templateLinkedPrincipal](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"entity_id": schema.StringAttribute{
													Required: true,
												},
												"entity_type": schema.StringAttribute{
													Required: true,
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
												},
												"entity_type": schema.StringAttribute{
													Required: true,
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

const (
	resourcePolicyIDPartsCount = 2
)

func (r *resourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var plan resourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &verifiedpermissions.CreatePolicyInput{}

	in.ClientToken = aws.String(id.UniqueId())
	in.PolicyStoreId = aws.String(plan.PolicyStoreID.ValueString())

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
			PolicyTemplateId: aws.String(templateLinked.PolicyTemplateID.ValueString()),
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

			value.Principal = &awstypes.EntityIdentifier{
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

	rID, err := interflex.FlattenResourceId(idParts, resourcePolicyIDPartsCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicy, plan.PolicyStoreID.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, rID)
	plan.CreatedDate = timetypes.NewRFC3339TimePointerValue(out.CreatedDate)
	plan.PolicyID = flex.StringToFramework(ctx, out.PolicyId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var state resourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rID, err := interflex.ExpandResourceId(state.ID.ValueString(), resourcePolicyIDPartsCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionSetting, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findPolicyByID(ctx, conn, rID[0], rID[1])
	if tfresource.NotFound(err) {
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

	if val, ok := out.Definition.(*awstypes.PolicyDefinitionDetailMemberStatic); ok && val != nil {
		static := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &staticPolicyDefinition{
			Statement:   flex.StringToFramework(ctx, val.Value.Statement),
			Description: flex.StringToFramework(ctx, val.Value.Description),
		})

		state.Definition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &policyDefinition{
			Static:         static,
			TemplateLinked: fwtypes.NewListNestedObjectValueOfNull[templateLinkedPolicyDefinition](ctx),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var plan, state resourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Definition.Equal(state.Definition) {

		in := &verifiedpermissions.UpdatePolicyInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdatePolicy(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var state resourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &verifiedpermissions.DeletePolicyInput{
		PolicyId:      aws.String(state.PolicyID.ValueString()),
		PolicyStoreId: aws.String(state.PolicyStoreID.ValueString()),
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

func (r *resourcePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findPolicyByID(ctx context.Context, conn *verifiedpermissions.Client, id, policyStoreId string) (*verifiedpermissions.GetPolicyOutput, error) {
	in := &verifiedpermissions.GetPolicyInput{
		PolicyId:      aws.String(id),
		PolicyStoreId: aws.String(policyStoreId),
	}

	out, err := conn.GetPolicy(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.PolicyId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourcePolicyData struct {
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
	Resource         fwtypes.SetNestedObjectValueOf[templateLinkedResource]   `tfsdk:"resource"`
}

type templateLinkedPrincipal struct {
	EntityID   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
}

type templateLinkedResource struct {
	EntityID   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
}
