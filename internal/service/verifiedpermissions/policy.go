// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Policy")
func newResourcePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicy{}

	return r, nil
}

const (
	ResNamePolicy = "Policy"
)

type resourcePolicy struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourcePolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_verifiedpermissions_policy"
}

func (r *resourcePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_date": schema.StringAttribute{
				Computed: true,
			},
			"last_updated_date": schema.StringAttribute{
				Computed: true,
			},
			"policy_id": framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Required: true,
			},
			"policy_type": schema.StringAttribute{
				Required: true, // TODO
			},
			"principal": schema.StringAttribute{
				Optional: true,
			},
			"resource": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"static": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"description": schema.StringAttribute{
										Required: true,
									},
									"statement": schema.StringAttribute{
										Required: true,
									},
								},
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(path.Expressions{
										path.MatchRoot("template_linked"),
									}...),
								},
							},
						},
						"template_linked": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"policy_template_id": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"principal": schema.SetNestedBlock{
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
									"resource": schema.SetNestedBlock{
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
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(path.Expressions{
										path.MatchRoot("static"),
									}...),
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var plan resourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &verifiedpermissions.CreatePolicyInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clientToken := id.UniqueId()
	in.ClientToken = aws.String(clientToken)

	out, err := conn.CreatePolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicy, clientToken, err),
			err.Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.PolicyId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var state resourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPolicyByID(ctx, conn, state.ID.ValueString())
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
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

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ComplexArgument.Equal(state.ComplexArgument) ||
		!plan.Type.Equal(state.Type) {

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
		PolicyId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeletePolicy(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionWaitingForDeletion, ResNamePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourcePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findPolicyByID(ctx context.Context, conn *verifiedpermissions.Client, id string) (*verifiedpermissions.GetPolicyOutput, error) {
	in := &verifiedpermissions.GetPolicyInput{
		PolicyId: aws.String(id),
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
	ARN             types.String   `tfsdk:"arn"`
	ComplexArgument types.List     `tfsdk:"complex_argument"`
	Description     types.String   `tfsdk:"description"`
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	Type            types.String   `tfsdk:"type"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
