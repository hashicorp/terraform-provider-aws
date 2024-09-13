// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceAccessPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAccessPolicy{}, nil
}

type resourceAccessPolicyData struct {
	Description   types.String                                  `tfsdk:"description"`
	ID            types.String                                  `tfsdk:"id"`
	Name          types.String                                  `tfsdk:"name"`
	Policy        jsontypes.Normalized                          `tfsdk:"policy"`
	PolicyVersion types.String                                  `tfsdk:"policy_version"`
	Type          fwtypes.StringEnum[awstypes.AccessPolicyType] `tfsdk:"type"`
}

const (
	ResNameAccessPolicy = "Access Policy"
)

type resourceAccessPolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceAccessPolicy) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_opensearchserverless_access_policy"
}

func (r *resourceAccessPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20480),
				},
			},
			"policy_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessPolicyType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceAccessPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAccessPolicyData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	in := &opensearchserverless.CreateAccessPolicyInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	in.ClientToken = aws.String(id.UniqueId())

	out, err := conn.CreateAccessPolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameAccessPolicy, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(flex.Flatten(ctx, out.AccessPolicyDetail, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = flex.StringToFramework(ctx, out.AccessPolicyDetail.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccessPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceAccessPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccessPolicyByNameAndType(ctx, conn, state.ID.ValueString(), state.Type.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, ResNameAccessPolicy, state.ID.ValueString(), err),
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

func (r *resourceAccessPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state resourceAccessPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.Policy.Equal(state.Policy) {
		input := &opensearchserverless.UpdateAccessPolicyInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(id.UniqueId())

		out, err := conn.UpdateAccessPolicy(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}
		resp.Diagnostics.Append(flex.Flatten(ctx, out.AccessPolicyDetail, &state)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccessPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceAccessPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteAccessPolicy(ctx, &opensearchserverless.DeleteAccessPolicyInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(state.Name.ValueString()),
		Type:        awstypes.AccessPolicyType(state.Type.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameAccessPolicy, state.Name.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceAccessPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, idSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected security-policy-name%[2]ssecurity-policy-type", req.ID, idSeparator)
		resp.Diagnostics.AddError(fmt.Sprintf("importing Security Policy (%s)", req.ID), err.Error())
		return
	}

	state := resourceAccessPolicyData{
		ID:   types.StringValue(parts[0]),
		Name: types.StringValue(parts[0]),
		Type: fwtypes.StringEnumValue(awstypes.AccessPolicyType(parts[1])),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
