// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssmquicksetup_configuration_manager", name="Configuration Manager")
// @Tags(identifierAttribute="manager_arn")
func newConfigurationManagerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &configurationManagerResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type configurationManagerResource struct {
	framework.ResourceWithModel[configurationManagerResourceModel]
	framework.WithTimeouts
}

func (r *configurationManagerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				// The API returns an empty string when description is omitted. To prevent "inconsistent
				// final plan" errors when null, mark this argument as optional/computed.
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"manager_arn": framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"status_summaries": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[statusSummaryModel](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[statusSummaryModel](ctx),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"configuration_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[configurationDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"local_deployment_administration_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
						"local_deployment_execution_role_name": schema.StringAttribute{
							Optional: true,
						},
						names.AttrParameters: schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							Required:    true,
							ElementType: types.StringType,
						},
						names.AttrType: schema.StringAttribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *configurationManagerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data configurationManagerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMQuickSetupClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input ssmquicksetup.CreateConfigurationManagerInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	outputCCM, err := conn.CreateConfigurationManager(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SSM Quick Setup Configuration Manager (%s)", name), err.Error())

		return
	}

	arn := aws.ToString(outputCCM.ManagerArn)
	data.ManagerARN = fwflex.StringValueToFramework(ctx, arn)

	outputGCM, err := waitConfigurationManagerCreated(ctx, conn, arn, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SSM Quick Setup Configuration Manager (%s) create", arn), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGCM, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *configurationManagerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data configurationManagerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMQuickSetupClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ManagerARN)
	output, err := findConfigurationManagerByID(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSM Quick Setup Configuration Manager (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *configurationManagerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old configurationManagerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMQuickSetupClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, new.ManagerARN)

	if !new.Description.Equal(old.Description) || !new.Name.Equal(old.Name) {
		var input ssmquicksetup.UpdateConfigurationManagerInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateConfigurationManager(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SSM Quick Setup Configuration Manager (%s)", arn), err.Error())

			return
		}
	}

	if !new.ConfigurationDefinition.Equal(old.ConfigurationDefinition) {
		var inputs []ssmquicksetup.UpdateConfigurationDefinitionInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new.ConfigurationDefinition, &inputs)...)
		if response.Diagnostics.HasError() {
			return
		}

		for _, input := range inputs {
			input.ManagerArn = aws.String(arn)

			_, err := conn.UpdateConfigurationDefinition(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating SSM Quick Setup Configuration Manager (%s)", arn), err.Error())

				return
			}
		}
	}

	output, err := waitConfigurationManagerUpdated(ctx, conn, arn, r.UpdateTimeout(ctx, new.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SSM Quick Setup Configuration Manager (%s) update", arn), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *configurationManagerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data configurationManagerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMQuickSetupClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ManagerARN)
	input := ssmquicksetup.DeleteConfigurationManagerInput{
		ManagerArn: aws.String(arn),
	}
	_, err := conn.DeleteConfigurationManager(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SSM Quick Setup Configuration Manager (%s)", arn), err.Error())

		return
	}

	if _, err := waitConfigurationManagerDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SSM Quick Setup Configuration Manager (%s) delete", arn), err.Error())

		return
	}
}

func (r *configurationManagerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("manager_arn"), request, response)
}

func findConfigurationManagerByID(ctx context.Context, conn *ssmquicksetup.Client, arn string) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	input := ssmquicksetup.GetConfigurationManagerInput{
		ManagerArn: aws.String(arn),
	}
	output, err := conn.GetConfigurationManager(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusConfigurationManager(conn *ssmquicksetup.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findConfigurationManagerByID(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// GetConfigurationManager returns an array of status summaries. The item
		// with a "Deployment" type will contain the status of the configuration
		// manager during create, update, and delete.
		for _, v := range output.StatusSummaries {
			if v.StatusType == awstypes.StatusTypeDeployment {
				return output, string(v.Status), nil
			}
		}

		return nil, "", nil
	}
}

func waitConfigurationManagerCreated(ctx context.Context, conn *ssmquicksetup.Client, arn string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInitializing, awstypes.StatusDeploying),
		Target:  enum.Slice(awstypes.StatusSucceeded),
		Refresh: statusConfigurationManager(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		for _, v := range output.StatusSummaries {
			if v.StatusType == awstypes.StatusTypeDeployment {
				retry.SetLastError(err, errors.New(aws.ToString(v.StatusMessage)))
			}
		}

		return output, err
	}

	return nil, err
}

func waitConfigurationManagerUpdated(ctx context.Context, conn *ssmquicksetup.Client, arn string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInitializing, awstypes.StatusDeploying),
		Target:  enum.Slice(awstypes.StatusSucceeded),
		Refresh: statusConfigurationManager(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		for _, v := range output.StatusSummaries {
			if v.StatusType == awstypes.StatusTypeDeployment {
				retry.SetLastError(err, errors.New(aws.ToString(v.StatusMessage)))
			}
		}

		return output, err
	}

	return nil, err
}

func waitConfigurationManagerDeleted(ctx context.Context, conn *ssmquicksetup.Client, arn string, timeout time.Duration) (*ssmquicksetup.GetConfigurationManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeploying, awstypes.StatusStopping, awstypes.StatusDeleting),
		Target:  []string{},
		Refresh: statusConfigurationManager(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssmquicksetup.GetConfigurationManagerOutput); ok {
		for _, v := range output.StatusSummaries {
			if v.StatusType == awstypes.StatusTypeDeployment {
				retry.SetLastError(err, errors.New(aws.ToString(v.StatusMessage)))
			}
		}

		return output, err
	}

	return nil, err
}

type configurationManagerResourceModel struct {
	framework.WithRegionModel
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
	ID                                   types.String        `tfsdk:"id"`
	LocalDeploymentAdministrationRoleARN fwtypes.ARN         `tfsdk:"local_deployment_administration_role_arn"`
	LocalDeploymentExecutionRoleName     types.String        `tfsdk:"local_deployment_execution_role_name"`
	Parameters                           fwtypes.MapOfString `tfsdk:"parameters"`
	Type                                 types.String        `tfsdk:"type"`
	TypeVersion                          types.String        `tfsdk:"type_version"`
}

type statusSummaryModel struct {
	Status        types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
	StatusType    types.String `tfsdk:"status_type"`
}
