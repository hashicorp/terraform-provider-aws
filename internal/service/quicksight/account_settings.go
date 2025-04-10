// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_account_settings", name="Account Settings")
func newResourceAccountSettings(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAccountSettings{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameAccountSettings = "Account Settings"
)

type resourceAccountSettings struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceAccountSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_namespace": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			"reset_on_delete": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				DeprecationMessage: `The "reset_on_delete" attribute will be removed in a future version of the provider`,
			},
			"termination_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},

		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAccountSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)
	awsAccountID := r.Meta().AccountID(ctx)
	var plan resourceAccountSettingsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := quicksight.UpdateAccountSettingsInput{
		AwsAccountId: &awsAccountID,
		// API currently does not support overriding default namespace, but requires for API call
		DefaultNamespace: aws.String("default"),
	}

	if !plan.TerminationProtectionEnabled.IsNull() {
		input.TerminationProtectionEnabled = plan.TerminationProtectionEnabled.ValueBool()
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := tfresource.RetryGWhen(ctx, createTimeout,
		func() (*quicksight.DescribeAccountSettingsOutput, error) {
			_, err := conn.UpdateAccountSettings(ctx, &input)
			if err != nil {
				return nil, err
			}

			input := quicksight.DescribeAccountSettingsInput{
				AwsAccountId: aws.String(awsAccountID),
			}

			return conn.DescribeAccountSettings(ctx, &input)
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
	if err != nil {
		resp.Diagnostics.AddError("creating Quicksight Account", err.Error())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output.AccountSettings, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(awsAccountID)
	plan.DefaultNamespace = types.StringValue("default")

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAccountSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)
	awsAccountID := r.Meta().AccountID(ctx)
	var state resourceAccountSettingsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccountSettingsByID(ctx, conn, awsAccountID)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameAccountSettings, state.ID.String(), err),
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

func (r *resourceAccountSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)
	var plan, state resourceAccountSettingsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		awsAccountID := r.Meta().AccountID(ctx)
		input := quicksight.UpdateAccountSettingsInput{
			AwsAccountId: &awsAccountID,
			// API currently does not support overriding default namespace, but requires for API call
			DefaultNamespace: aws.String("default"),
		}
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.TerminationProtectionEnabled.IsNull() {
			input.TerminationProtectionEnabled = plan.TerminationProtectionEnabled.ValueBool()
		}

		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		output, err := tfresource.RetryGWhen(ctx, createTimeout,
			func() (*quicksight.DescribeAccountSettingsOutput, error) {
				_, err := conn.UpdateAccountSettings(ctx, &input)
				if err != nil {
					return nil, err
				}

				input := quicksight.DescribeAccountSettingsInput{
					AwsAccountId: aws.String(awsAccountID),
				}

				return conn.DescribeAccountSettings(ctx, &input)
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
		if err != nil {
			resp.Diagnostics.AddError("creating Quicksight Account", err.Error())
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, output.AccountSettings, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAccountSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceAccountSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ResetOnDelete.ValueBool() {
		conn := r.Meta().QuickSightClient(ctx)
		awsAccountID := r.Meta().AccountID(ctx)
		input := quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 &awsAccountID,
			TerminationProtectionEnabled: true,
			DefaultNamespace:             aws.String("default"),
		}

		_, err := conn.UpdateAccountSettings(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError("resetting Quicksight Account Settings", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddWarning(
			"Resource Destruction",
			"This resource has only been removed from Terraform state. "+
				"Manually use the AWS Console to fully destroy this resource. "+
				"Setting the attribute \"reset_on_delete\" will also fully destroy resources of this type.",
		)
	}
}

func (r *resourceAccountSettings) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		var resetOnDelete types.Bool
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("reset_on_delete"), &resetOnDelete)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !resetOnDelete.ValueBool() {
			resp.Diagnostics.AddWarning(
				"Resource Destruction",
				"Applying this resource destruction will only remove the resource from Terraform state and will not reset account settings. "+
					"Either manually use the AWS Console to fully destroy this resource or "+
					"update the resource with \"reset_on_delete\" set to true.",
			)
		}
	}
}

func (r *resourceAccountSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findAccountSettingsByID(ctx context.Context, conn *quicksight.Client, id string) (*awstypes.AccountSettings, error) {
	input := quicksight.DescribeAccountSettingsInput{
		AwsAccountId: aws.String(id),
	}

	out, err := conn.DescribeAccountSettings(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.AccountSettings == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.AccountSettings, nil
}

type resourceAccountSettingsModel struct {
	DefaultNamespace             types.String   `tfsdk:"default_namespace"`
	ID                           types.String   `tfsdk:"id"`
	ResetOnDelete                types.Bool     `tfsdk:"reset_on_delete"`
	TerminationProtectionEnabled types.Bool     `tfsdk:"termination_protection_enabled"`
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
}
