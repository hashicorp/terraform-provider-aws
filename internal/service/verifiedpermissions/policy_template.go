// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Policy Template")
func newResourcePolicyTemplate(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyTemplate{}

	return r, nil
}

const (
	ResNamePolicyTemplate = "Policy Template"
)

type resourcePolicyTemplate struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourcePolicyTemplate) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_verifiedpermissions_policy_template"
}

// Schema returns the schema for this resource.
func (r *resourcePolicyTemplate) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_template_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"statement": schema.StringAttribute{
				Required: true,
			},
		},
	}

	response.Schema = s
}

func (r *resourcePolicyTemplate) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var plan resourcePolicyTemplateData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &verifiedpermissions.CreatePolicyTemplateInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())

	output, err := conn.CreatePolicyTemplate(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyTemplate, plan.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringValueToFramework(ctx, fmt.Sprintf("%s:%s", aws.ToString(output.PolicyStoreId), aws.ToString(output.PolicyTemplateId)))

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourcePolicyTemplate) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourcePolicyTemplateData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	policyStoreID, policyTemplateID, err := policyTemplateParseID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyTemplate, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	output, err := findPolicyTemplateByID(ctx, conn, policyStoreID, policyTemplateID)

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyTemplate, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourcePolicyTemplate) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state, plan resourcePolicyTemplateData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) || !plan.Statement.Equal(state.Statement) {
		input := &verifiedpermissions.UpdatePolicyTemplateInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)

		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdatePolicyTemplate(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyTemplate, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourcePolicyTemplate) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourcePolicyTemplateData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Verified Permissions Policy Template", map[string]interface{}{
		names.AttrID: state.ID.ValueString(),
	})

	input := &verifiedpermissions.DeletePolicyTemplateInput{
		PolicyStoreId:    aws.String(state.PolicyStoreID.ValueString()),
		PolicyTemplateId: aws.String(state.PolicyTemplateID.ValueString()),
	}

	_, err := conn.DeletePolicyTemplate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyTemplate, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

type resourcePolicyTemplateData struct {
	CreatedDate      timetypes.RFC3339 `tfsdk:"created_date"`
	Description      types.String      `tfsdk:"description"`
	ID               types.String      `tfsdk:"id"`
	PolicyStoreID    types.String      `tfsdk:"policy_store_id"`
	PolicyTemplateID types.String      `tfsdk:"policy_template_id"`
	Statement        types.String      `tfsdk:"statement"`
}

func findPolicyTemplateByID(ctx context.Context, conn *verifiedpermissions.Client, policyStoreId, id string) (*verifiedpermissions.GetPolicyTemplateOutput, error) {
	in := &verifiedpermissions.GetPolicyTemplateInput{
		PolicyStoreId:    aws.String(policyStoreId),
		PolicyTemplateId: aws.String(id),
	}
	out, err := conn.GetPolicyTemplate(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.PolicyStoreId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func policyTemplateParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%s), expected POLICY-STORE-ID:POLICY-TEMPLATE-ID", id)
}
