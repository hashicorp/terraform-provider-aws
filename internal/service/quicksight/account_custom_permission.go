// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_account_custom_permission", name="Account Custom Permission")
// @IdentityAttribute("aws_account_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
func newAccountCustomPermissionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountCustomPermissionResource{}

	return r, nil
}

type accountCustomPermissionResource struct {
	framework.ResourceWithModel[accountCustomPermissionResourceModel]
	framework.WithImportByIdentity
}

func (r *accountCustomPermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"custom_permissions_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *accountCustomPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var data accountCustomPermissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	var input quicksight.UpdateAccountCustomPermissionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateAccountCustomPermission(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.AWSAccountID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *accountCustomPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var data accountCustomPermissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccountCustomPermissionByID(ctx, conn, data.AWSAccountID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.AWSAccountID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *accountCustomPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan, state accountCustomPermissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input quicksight.UpdateAccountCustomPermissionInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAccountCustomPermission(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AWSAccountID.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accountCustomPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state accountCustomPermissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	var input quicksight.DeleteAccountCustomPermissionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteAccountCustomPermission(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AWSAccountID.ValueString())
		return
	}
}

func findAccountCustomPermissionByID(ctx context.Context, conn *quicksight.Client, accountID string) (*quicksight.DescribeAccountCustomPermissionOutput, error) {
	input := quicksight.DescribeAccountCustomPermissionInput{
		AwsAccountId: aws.String(accountID),
	}

	out, err := conn.DescribeAccountCustomPermission(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || aws.ToString(out.CustomPermissionsName) == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type accountCustomPermissionResourceModel struct {
	framework.WithRegionModel
	AWSAccountID          types.String `tfsdk:"aws_account_id"`
	CustomPermissionsName types.String `tfsdk:"custom_permissions_name"`
}
