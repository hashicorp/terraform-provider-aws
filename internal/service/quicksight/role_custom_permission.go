// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package quicksight

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_role_custom_permission", name="Role Custom Permission")
func newRoleCustomPermissionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &roleCustomPermissionResource{}

	return r, nil
}

type roleCustomPermissionResource struct {
	framework.ResourceWithModel[roleCustomPermissionResourceModel]
}

func (r *roleCustomPermissionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"custom_permissions_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrNamespace: quicksightschema.NamespaceAttribute(),
			names.AttrRole: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Role](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *roleCustomPermissionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data roleCustomPermissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateRoleCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateRoleCustomPermission(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Role (%s) Custom Permission (%s)", data.Role.ValueString(), data.CustomPermissionsName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *roleCustomPermissionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data roleCustomPermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	output, err := findRoleCustomPermissionByThreePartKey(ctx, conn, data.AWSAccountID.ValueString(), data.Namespace.ValueString(), data.Role.ValueEnum())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Role (%s) Custom Permission", data.Role.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.CustomPermissionsName = fwflex.StringToFramework(ctx, output)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *roleCustomPermissionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old roleCustomPermissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateRoleCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateRoleCustomPermission(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Quicksight Role (%s) Custom Permission (%s)", new.Role.ValueString(), new.CustomPermissionsName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *roleCustomPermissionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data roleCustomPermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.DeleteRoleCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteRoleCustomPermission(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight Role (%s) Custom Permission (%s)", data.Role.ValueString(), data.CustomPermissionsName.ValueString()), err.Error())

		return
	}
}

func (r *roleCustomPermissionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		roleCustomPermissionIDParts = 3
	)
	parts, err := intflex.ExpandResourceId(request.ID, roleCustomPermissionIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAWSAccountID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrNamespace), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRole), parts[2])...)
}

func findRoleCustomPermissionByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string, role awstypes.Role) (*string, error) {
	input := quicksight.DescribeRoleCustomPermissionInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		Role:         role,
	}

	return findRoleCustomPermission(ctx, conn, &input)
}

func findRoleCustomPermission(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeRoleCustomPermissionInput) (*string, error) {
	output, err := conn.DescribeRoleCustomPermission(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || aws.ToString(output.CustomPermissionsName) == "" {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.CustomPermissionsName, nil
}

type roleCustomPermissionResourceModel struct {
	framework.WithRegionModel
	AWSAccountID          types.String                      `tfsdk:"aws_account_id"`
	CustomPermissionsName types.String                      `tfsdk:"custom_permissions_name"`
	Namespace             types.String                      `tfsdk:"namespace"`
	Role                  fwtypes.StringEnum[awstypes.Role] `tfsdk:"role"`
}
