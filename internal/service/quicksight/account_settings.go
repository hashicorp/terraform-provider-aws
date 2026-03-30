// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package quicksight

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_account_settings", name="Account Settings")
func newAccountSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountSettingsResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type accountSettingsResource struct {
	framework.ResourceWithModel[accountSettingsResourceModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *accountSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"default_namespace":    quicksightschema.NamespaceAttribute(),
			"termination_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *accountSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateAccountSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := updateAccountSettings(ctx, conn, &input, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Account Settings (%s)", data.AWSAccountID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *accountSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	output, err := findAccountSettingsByID(ctx, conn, data.AWSAccountID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Quicksight Account Settings (%s)", data.AWSAccountID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old accountSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.UpdateAccountSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := updateAccountSettings(ctx, conn, &input, r.UpdateTimeout(ctx, new.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Quicksight Account Settings (%s)", new.AWSAccountID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accountSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrAWSAccountID), request, response)
}

func updateAccountSettings(ctx context.Context, conn *quicksight.Client, input *quicksight.UpdateAccountSettingsInput, timeout time.Duration) (*awstypes.AccountSettings, error) {
	return tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (*awstypes.AccountSettings, error) {
			_, err := conn.UpdateAccountSettings(ctx, input)

			if err != nil {
				return nil, err
			}

			return findAccountSettingsByID(ctx, conn, aws.ToString(input.AwsAccountId))
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "You don't have access to this item.\n  The provided credentials couldn't be validated.\n  You might not be authorized to carry out the request.\n  Make sure that your account is authorized to use the Amazon QuickSight service, that your policies have the correct permissions, and that you are using the correct credentials.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InternalFailureException](err, "An internal failure occurred.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "One or more parameters has a value that isn't valid.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "One or more resources can't be found.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ResourceUnavailableException](err, "This resource is currently unavailable.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ThrottlingException](err, "Access is throttled.") {
				return true, err
			}
			return false, err
		},
	)
}

func findAccountSettingsByID(ctx context.Context, conn *quicksight.Client, id string) (*awstypes.AccountSettings, error) {
	input := quicksight.DescribeAccountSettingsInput{
		AwsAccountId: aws.String(id),
	}
	output, err := conn.DescribeAccountSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccountSettings == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.AccountSettings, nil
}

type accountSettingsResourceModel struct {
	framework.WithRegionModel
	AWSAccountID                 types.String   `tfsdk:"aws_account_id"`
	DefaultNamespace             types.String   `tfsdk:"default_namespace"`
	TerminationProtectionEnabled types.Bool     `tfsdk:"termination_protection_enabled"`
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
}
