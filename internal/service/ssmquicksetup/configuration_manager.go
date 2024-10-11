// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmquicksetup/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssmquicksetup_configuration_manager", name="Configuration Manager")
// @Tags(identifierAttribute="manager_arn")
func newResourceConfigurationManager(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceConfigurationManager{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameConfigurationManager = "Configuration Manager"
)

type resourceConfigurationManager struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceConfigurationManager) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssmquicksetup_configuration_manager"
}

func (r *resourceConfigurationManager) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional: true,
			},
			"manager_arn": framework.ARNAttributeComputedOnly(),
			"name": schema.StringAttribute{
				Required: true,
			},
			"status_summaries": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[statusSummaryModel](ctx),
				ElementType: fwtypes.NewObjectTypeOf[statusSummaryModel](ctx),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"configuration_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[configurationDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"local_deployment_administration_role_arn": schema.StringAttribute{
							Optional: true,
						},
						"local_deployment_execution_role_name": schema.StringAttribute{
							Optional: true,
						},
						"parameters": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"type_version": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
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

func (r *resourceConfigurationManager) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSMQuickSetupClient(ctx)

	var plan resourceConfigurationManagerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ssmquicksetup.CreateConfigurationManagerInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateConfigurationManager(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionCreating, ResNameConfigurationManager, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ManagerArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionCreating, ResNameConfigurationManager, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ManagerARN = flex.StringToFramework(ctx, out.ManagerArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	statusOut, err := waitConfigurationManagerCreated(ctx, conn, plan.ManagerARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionWaitingForCreation, ResNameConfigurationManager, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, statusOut, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceConfigurationManager) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSMQuickSetupClient(ctx)

	var state resourceConfigurationManagerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findConfigurationManagerByID(ctx, conn, state.ManagerARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionSetting, ResNameConfigurationManager, state.ManagerARN.String(), err),
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

func (r *resourceConfigurationManager) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSMQuickSetupClient(ctx)

	var plan, state resourceConfigurationManagerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) {
		var input ssmquicksetup.UpdateConfigurationManagerInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateConfigurationManager(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionUpdating, ResNameConfigurationManager, plan.ManagerARN.String(), err),
				err.Error(),
			)
			return
		}
	}

	if !plan.ConfigurationDefinition.Equal(state.ConfigurationDefinition) {
		var inputs []ssmquicksetup.UpdateConfigurationDefinitionInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan.ConfigurationDefinition, &inputs)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, input := range inputs {
			input.ManagerArn = plan.ManagerARN.ValueStringPointer()

			_, err := conn.UpdateConfigurationDefinition(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionUpdating, ResNameConfigurationManager, plan.ManagerARN.String(), err),
					err.Error(),
				)
				return
			}
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	statusOut, err := waitConfigurationManagerUpdated(ctx, conn, plan.ManagerARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionWaitingForUpdate, ResNameConfigurationManager, plan.ManagerARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, statusOut, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceConfigurationManager) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSMQuickSetupClient(ctx)

	var state resourceConfigurationManagerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ssmquicksetup.DeleteConfigurationManagerInput{
		ManagerArn: state.ManagerARN.ValueStringPointer(),
	}

	_, err := conn.DeleteConfigurationManager(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionDeleting, ResNameConfigurationManager, state.ManagerARN.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitConfigurationManagerDeleted(ctx, conn, state.ManagerARN.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMQuickSetup, create.ErrActionWaitingForDeletion, ResNameConfigurationManager, state.ManagerARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceConfigurationManager) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("manager_arn"), req, resp)
}

func (r *resourceConfigurationManager) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitConfigurationManagerCreated(ctx context.Context, conn *ssmquicksetup.Client, id string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInitializing, awstypes.StatusDeploying),
		Target:  enum.Slice(awstypes.StatusSucceeded),
		Refresh: statusConfigurationManager(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		return out, err
	}

	return nil, err
}

func waitConfigurationManagerUpdated(ctx context.Context, conn *ssmquicksetup.Client, id string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInitializing, awstypes.StatusDeploying),
		Target:  enum.Slice(awstypes.StatusSucceeded),
		Refresh: statusConfigurationManager(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		return out, err
	}

	return nil, err
}

func waitConfigurationManagerDeleted(ctx context.Context, conn *ssmquicksetup.Client, id string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeploying, awstypes.StatusStopping, awstypes.StatusDeleting),
		Target:  []string{},
		Refresh: statusConfigurationManager(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		return out, err
	}

	return nil, err
}

func statusConfigurationManager(ctx context.Context, conn *ssmquicksetup.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findConfigurationManagerByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		// GetConfigurationManager returns an array of status summaries. The item
		// with a "Deployment" type will contain the status of the configuration
		// manager during create, update, and delete.
		for _, st := range out.StatusSummaries {
			if st.StatusType == awstypes.StatusTypeDeployment {
				return out, string(st.Status), nil
			}
		}

		return out, "", nil
	}
}

func findConfigurationManagerByID(ctx context.Context, conn *ssmquicksetup.Client, id string) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	in := &ssmquicksetup.GetConfigurationManagerInput{
		ManagerArn: aws.String(id),
	}

	out, err := conn.GetConfigurationManager(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceConfigurationManagerModel struct {
	ConfigurationDefinition fwtypes.ListNestedObjectValueOf[configurationDefinitionModel] `tfsdk:"configuration_definition"`
	Description             types.String                                                  `tfsdk:"description"`
	ManagerARN              types.String                                                  `tfsdk:"manager_arn"`
	Name                    types.String                                                  `tfsdk:"name"`
	StatusSummaries         fwtypes.ListNestedObjectValueOf[statusSummaryModel]           `tfsdk:"status_summaries"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

type configurationDefinitionModel struct {
	ID                                   types.String                     `tfsdk:"id"`
	LocalDeploymentAdministrationRoleARN types.String                     `tfsdk:"local_deployment_administration_role_arn"`
	LocalDeploymentExecutionRoleName     types.String                     `tfsdk:"local_deployment_execution_role_name"`
	Parameters                           fwtypes.MapValueOf[types.String] `tfsdk:"parameters"`
	Type                                 types.String                     `tfsdk:"type"`
	TypeVersion                          types.String                     `tfsdk:"type_version"`
}

type statusSummaryModel struct {
	Status        types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
	StatusType    types.String `tfsdk:"status_type"`
}
