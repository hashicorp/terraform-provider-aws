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

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().RAMClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	input := ram.AssociateResourceSharePermissionInput{
		ClientToken:      aws.String(create.UniqueId(ctx)),
		PermissionArn:    aws.String(plan.PermissionARN.ValueString()),
		ResourceShareArn: aws.String(plan.ResourceShareARN.ValueString()),
		Replace:          aws.Bool(true),
	}
	// TIP: -- 4. Call the AWS Create function
	_, err := conn.AssociateResourceSharePermission(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionCreating, ResNameResourceSharePermissionAssociation, plan.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}
	// TIP: -- 5. Set minimum attributes for Read to work
	// No waiter needed — API returns boolean instantly.
	// Set the composite ID so Read can find the resource.
	plan.ResourceShareARN = types.StringValue(plan.ResourceShareARN.ValueString())
	plan.PermissionARN = types.StringValue(plan.PermissionARN.ValueString())

	// TIP: -- 6. No waiter needed for this resource.
	// AssociateResourceSharePermission returns returnValue: boolean immediately.
	// Unlike AssociateResourceShare which returns an association with async status,
	// this API completes synchronously.
	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSharePermissionAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().RAMClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS
	// We use ListResourceSharePermissions and filter by permissionArn
	// since there is no direct Get API for permission associations.
	out, err := findResourceSharePermissionByTwoPartKey(ctx, conn, state.ResourceShareARN.ValueString(), state.PermissionARN.ValueString())

	// TIP: -- 4. Remove resource from state if it is not found
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionReading, ResNameResourceSharePermissionAssociation, state.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	// Set all schema attributes from the AWS response.
	// permission_version is Computed so we read it back from AWS here.
	state.PermissionARN = types.StringValue(aws.ToString(out.Arn))
	state.ResourceShareARN = types.StringValue(state.ResourceShareARN.ValueString())
	if v, err := strconv.ParseInt(aws.ToString(out.Version), 10, 64); err == nil {
		state.PermissionVersion = types.Int64Value(v)
	}

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSharePermissionAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().RAMClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceSharePermissionAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := ram.DisassociateResourceSharePermissionInput{
		ClientToken:      aws.String(create.UniqueId(ctx)),
		PermissionArn:    aws.String(state.PermissionARN.ValueString()),
		ResourceShareArn: aws.String(state.ResourceShareARN.ValueString()),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DisassociateResourceSharePermission(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RAM, create.ErrActionDeleting, ResNameResourceSharePermissionAssociation, state.PermissionARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. No waiter needed.
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
