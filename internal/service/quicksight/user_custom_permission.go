// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

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
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_user_custom_permission", name="User Custom Permission")
func newUserCustomPermissionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &userCustomPermissionResource{}

	return r, nil
}

type userCustomPermissionResource struct {
	framework.ResourceWithModel[userCustomPermissionResourceModel]
}

func (r *userCustomPermissionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"custom_permissions_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrNamespace: quicksightschema.NamespaceAttribute(),
			names.AttrUserName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userCustomPermissionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userCustomPermissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateUserCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateUserCustomPermission(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight User (%s) Custom Permission (%s)", data.UserName.ValueString(), data.CustomPermissionsName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userCustomPermissionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userCustomPermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	output, err := findUserCustomPermissionByThreePartKey(ctx, conn, data.AWSAccountID.ValueString(), data.Namespace.ValueString(), data.UserName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight User (%s) Custom Permission", data.UserName.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.CustomPermissionsName = fwflex.StringToFramework(ctx, output)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userCustomPermissionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old userCustomPermissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateUserCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateUserCustomPermission(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Quicksight User (%s) Custom Permission (%s)", new.UserName.ValueString(), new.CustomPermissionsName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *userCustomPermissionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userCustomPermissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.DeleteUserCustomPermissionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteUserCustomPermission(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight User (%s) Custom Permission (%s)", data.UserName.ValueString(), data.CustomPermissionsName.ValueString()), err.Error())

		return
	}
}

func (r *userCustomPermissionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		userCustomPermissionIDParts = 3
	)
	parts, err := intflex.ExpandResourceId(request.ID, userCustomPermissionIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAWSAccountID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrNamespace), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrUserName), parts[2])...)
}

func findUserCustomPermissionByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, userName string) (*string, error) {
	output, err := findUserByThreePartKey(ctx, conn, awsAccountID, namespace, userName)

	if err != nil {
		return nil, err
	}

	if aws.ToString(output.CustomPermissionsName) == "" {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.CustomPermissionsName, nil
}

type userCustomPermissionResourceModel struct {
	framework.WithRegionModel
	AWSAccountID          types.String `tfsdk:"aws_account_id"`
	CustomPermissionsName types.String `tfsdk:"custom_permissions_name"`
	Namespace             types.String `tfsdk:"namespace"`
	UserName              types.String `tfsdk:"user_name"`
}
