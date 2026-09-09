// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mwaaserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mwaaserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mwaaserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_mwaaserverless_workflow", name="Workflow")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mwaaserverless;mwaaserverless.GetWorkflowOutput")
func newWorkflowResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &workflowResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameWorkflow = "Workflow"
)

type workflowResource struct {
	framework.ResourceWithModel[workflowResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *workflowResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrEngineVersion: schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.WorkflowStatus](),
				Computed:   true,
			},
			"trigger_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_definition": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"definition_s3_location": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3LocationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrBucket: schema.StringAttribute{
							Required: true,
						},
						"object_key": schema.StringAttribute{
							Required: true,
						},
						"version_id": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EncryptionType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrLoggingConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[loggingConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrLogGroupName: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
						},
					},
				},
			},
			"schedule_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduleConfigurationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cron_expression": schema.StringAttribute{
							Computed: true,
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

func (r *workflowResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan workflowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MWAAServerlessClient(ctx)

	var input mwaaserverless.CreateWorkflowInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateWorkflow(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionCreating, ResNameWorkflow, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.WorkflowArn)
	plan.ARN = fwflex.StringValueToFramework(ctx, arn)
	plan.ID = fwflex.StringValueToFramework(ctx, arn)

	workflow, err := waitWorkflowCreated(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), arn)...)
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionWaitingForCreation, ResNameWorkflow, arn, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, workflow, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state workflowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MWAAServerlessClient(ctx)

	output, err := findWorkflowByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionReading, ResNameWorkflow, state.ID.ValueString(), err),
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

func (r *workflowResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state workflowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MWAAServerlessClient(ctx)

	// Only the arguments accepted by UpdateWorkflow are considered here. Changes to
	// "name" and "encryption_configuration" force a new resource. Tag-only changes
	// are handled by the transparent tagging interceptor.
	if !plan.DefinitionS3Location.Equal(state.DefinitionS3Location) ||
		!plan.RoleARN.Equal(state.RoleARN) ||
		!plan.Description.Equal(state.Description) ||
		!plan.EngineVersion.Equal(state.EngineVersion) ||
		!plan.LoggingConfiguration.Equal(state.LoggingConfiguration) ||
		!plan.NetworkConfiguration.Equal(state.NetworkConfiguration) ||
		!plan.TriggerMode.Equal(state.TriggerMode) {
		var input mwaaserverless.UpdateWorkflowInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.WorkflowArn = plan.ID.ValueStringPointer()

		_, err := conn.UpdateWorkflow(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionUpdating, ResNameWorkflow, plan.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		output, err := findWorkflowByARN(ctx, conn, plan.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionReading, ResNameWorkflow, plan.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state workflowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MWAAServerlessClient(ctx)

	input := mwaaserverless.DeleteWorkflowInput{
		WorkflowArn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteWorkflow(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionDeleting, ResNameWorkflow, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitWorkflowDeleted(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MWAAServerless, create.ErrActionWaitingForDeletion, ResNameWorkflow, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func findWorkflowByARN(ctx context.Context, conn *mwaaserverless.Client, arn string) (*mwaaserverless.GetWorkflowOutput, error) {
	input := mwaaserverless.GetWorkflowInput{
		WorkflowArn: aws.String(arn),
	}

	output, err := conn.GetWorkflow(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusWorkflow(conn *mwaaserverless.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findWorkflowByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.WorkflowStatus), nil
	}
}

func waitWorkflowCreated(ctx context.Context, conn *mwaaserverless.Client, arn string, timeout time.Duration) (*mwaaserverless.GetWorkflowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Target:                    enum.Slice(awstypes.WorkflowStatusReady),
		Refresh:                   statusWorkflow(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*mwaaserverless.GetWorkflowOutput); ok {
		return output, err
	}

	return nil, err
}

func waitWorkflowDeleted(ctx context.Context, conn *mwaaserverless.Client, arn string, timeout time.Duration) (*mwaaserverless.GetWorkflowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkflowStatusReady, awstypes.WorkflowStatusDeleting),
		Target:  []string{},
		Refresh: statusWorkflow(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*mwaaserverless.GetWorkflowOutput); ok {
		return output, err
	}

	return nil, err
}

type workflowResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	DefinitionS3Location    fwtypes.ListNestedObjectValueOf[s3LocationModel]              `tfsdk:"definition_s3_location"`
	Description             types.String                                                  `tfsdk:"description"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	EngineVersion           types.Int32                                                   `tfsdk:"engine_version"`
	ID                      types.String                                                  `tfsdk:"id"`
	LoggingConfiguration    fwtypes.ListNestedObjectValueOf[loggingConfigurationModel]    `tfsdk:"logging_configuration"`
	Name                    types.String                                                  `tfsdk:"name"`
	NetworkConfiguration    fwtypes.ListNestedObjectValueOf[networkConfigurationModel]    `tfsdk:"network_configuration"`
	RoleARN                 fwtypes.ARN                                                   `tfsdk:"role_arn"`
	ScheduleConfiguration   fwtypes.ListNestedObjectValueOf[scheduleConfigurationModel]   `tfsdk:"schedule_configuration"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
	TriggerMode             types.String                                                  `tfsdk:"trigger_mode"`
	WorkflowDefinition      types.String                                                  `tfsdk:"workflow_definition"`
	WorkflowStatus          fwtypes.StringEnum[awstypes.WorkflowStatus]                   `tfsdk:"status"`
	WorkflowVersion         types.String                                                  `tfsdk:"workflow_version"`
}

type s3LocationModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	ObjectKey types.String `tfsdk:"object_key"`
	VersionID types.String `tfsdk:"version_id"`
}

type encryptionConfigurationModel struct {
	Type     fwtypes.StringEnum[awstypes.EncryptionType] `tfsdk:"type"`
	KmsKeyID types.String                                `tfsdk:"kms_key_id"`
}

type loggingConfigurationModel struct {
	LogGroupName types.String `tfsdk:"log_group_name"`
}

type networkConfigurationModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
}

type scheduleConfigurationModel struct {
	CronExpression types.String `tfsdk:"cron_expression"`
}
