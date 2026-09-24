// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ram_resource_share_permission_association", name="Resource Share Permission Association")
// @IdentityAttribute("resource_share_arn")
// @IdentityAttribute("permission_arn")
// @Testing(hasNoPreExistingResource=true)
// @ImportIDHandler("resourceSharePermissionAssociationImportID")
func newResourceSharePermissionAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSharePermissionAssociationResource{}
	return r, nil
}

const (
	ResNameResourceSharePermissionAssociation = "Resource Share Permission Association"
)

type resourceSharePermissionAssociationResource struct {
	framework.ResourceWithModel[resourceSharePermissionAssociationResourceModel]
	framework.WithImportByIdentity
}

func (r *resourceSharePermissionAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associates an AWS RAM permission with a resource share.",
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"permission_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ARN of the RAM permission to associate with the resource share.",
			},
			"permission_version": schema.Int64Attribute{
				Computed:    true,
				Description: "The version of the RAM permission to associate with the resource share.",
			},
			"resource_share_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ARN of the RAM permission to associate with the resource share.",
			},
			// replace is intentionally not exposed as a schema attribute.
			// A resource share can only have one permission per resource type,
			// so replace must always be true when associating a permission.
			// See https://docs.aws.amazon.com/ram/latest/APIReference/API_AssociateResourceSharePermission.html
		},
	}
}

func (r *resourceSharePermissionAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().RAMClient(ctx)

	var plan resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ram.AssociateResourceSharePermissionInput{
		ClientToken:      aws.String(create.UniqueId(ctx)),
		PermissionArn:    aws.String(plan.PermissionARN.ValueString()),
		ResourceShareArn: aws.String(plan.ResourceShareARN.ValueString()),
		Replace:          aws.Bool(true),
	}

	_, err := conn.AssociateResourceSharePermission(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionCreating, ResNameResourceSharePermissionAssociation, plan.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.ResourceShareARN = types.StringValue(plan.ResourceShareARN.ValueString())
	plan.PermissionARN = types.StringValue(plan.PermissionARN.ValueString())
	plan.ID = types.StringValue(fmt.Sprintf("%s,%s", plan.ResourceShareARN.ValueString(), plan.PermissionARN.ValueString()))

	out, err := findResourceSharePermissionByTwoPartKey(ctx, conn, plan.ResourceShareARN.ValueString(), plan.PermissionARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionCreating, ResNameResourceSharePermissionAssociation, plan.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}
	if v, err := strconv.ParseInt(aws.ToString(out.Version), 10, 64); err == nil {
		plan.PermissionVersion = types.Int64Value(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSharePermissionAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().RAMClient(ctx)

	var state resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resourceShareARN, permissionARN string

	// Handle identity-based import where ID may not be set yet
	if state.ID.IsNull() || state.ID.ValueString() == "" {
		resourceShareARN = state.ResourceShareARN.ValueString()
		permissionARN = state.PermissionARN.ValueString()
	} else {
		parts := strings.SplitN(state.ID.ValueString(), ",", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			resp.Diagnostics.AddError(
				"Invalid ID",
				fmt.Sprintf("unexpected format of ID (%s), expected resource_share_arn,permission_arn", state.ID.ValueString()),
			)
			return
		}
		resourceShareARN, permissionARN = parts[0], parts[1]
	}

	out, err := findResourceSharePermissionByTwoPartKey(ctx, conn, resourceShareARN, permissionARN)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionReading, ResNameResourceSharePermissionAssociation, permissionARN, err),
			err.Error(),
		)
		return
	}

	state.PermissionARN = types.StringValue(aws.ToString(out.Arn))
	state.ResourceShareARN = types.StringValue(resourceShareARN)
	state.ID = types.StringValue(fmt.Sprintf("%s,%s", resourceShareARN, permissionARN))

	if v, err := strconv.ParseInt(aws.ToString(out.Version), 10, 64); err == nil {
		state.PermissionVersion = types.Int64Value(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSharePermissionAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().RAMClient(ctx)

	var state resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ram.DisassociateResourceSharePermissionInput{
		ClientToken:      aws.String(create.UniqueId(ctx)),
		PermissionArn:    aws.String(state.PermissionARN.ValueString()),
		ResourceShareArn: aws.String(state.ResourceShareARN.ValueString()),
	}

	_, err := conn.DisassociateResourceSharePermission(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return
		}
		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "already been disassociated") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionDeleting, ResNameResourceSharePermissionAssociation, state.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	// DisassociateResourceSharePermission returns returnValue: boolean instantly.
	// No async status to poll unlike DisassociateResourceShare.
}

func findResourceSharePermissionByTwoPartKey(ctx context.Context, conn *ram.Client, resourceShareARN, permissionARN string) (*awstypes.ResourceSharePermissionSummary, error) {
	input := ram.ListResourceSharePermissionsInput{
		ResourceShareArn: aws.String(resourceShareARN),
	}

	output, err := conn.ListResourceSharePermissions(ctx, &input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Permissions) == 0 {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("RAM Resource Share Permission Association (%s, %s) not found", resourceShareARN, permissionARN),
		}
	}

	// Filter the results to find the specific permission ARN
	for _, p := range output.Permissions {
		if aws.ToString(p.Arn) == permissionARN {
			return &p, nil
		}
	}

	// Permission ARN not found in the list
	return nil, &retry.NotFoundError{
		Message: fmt.Sprintf("RAM Resource Share Permission Association (%s, %s) not found", resourceShareARN, permissionARN),
	}
}

type resourceSharePermissionAssociationResourceModel struct {
	framework.WithRegionModel
	ID                types.String `tfsdk:"id"`
	PermissionARN     types.String `tfsdk:"permission_arn"`
	PermissionVersion types.Int64  `tfsdk:"permission_version"`
	ResourceShareARN  types.String `tfsdk:"resource_share_arn"`
}

type resourceSharePermissionAssociationImportID struct{}

func (resourceSharePermissionAssociationImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("unexpected format of ID (%s), expected resource_share_arn,permission_arn", id)
	}

	result := map[string]any{
		"resource_share_arn": parts[0],
		"permission_arn":     parts[1],
	}

	return id, result, nil
}

func (r *resourceSharePermissionAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	if req.ID == "" {
		r.WithImportByIdentity.ImportState(ctx, req, resp)
		return
	}

	parts := strings.SplitN(req.ID, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("unexpected format of ID (%s), expected resource_share_arn,permission_arn", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_share_arn"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("permission_arn"), parts[1])...)
}
