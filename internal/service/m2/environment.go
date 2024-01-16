// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="M2 Environment")
// @Tags(identifierAttribute="arn")
func newResourceEnvironment(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnvironment{}
	r.SetDefaultCreateTimeout(40 * time.Minute)
	r.SetDefaultUpdateTimeout(80 * time.Minute)
	r.SetDefaultDeleteTimeout(40 * time.Minute)

	return r, nil
}

const (
	ResNameEnvironment = "M2 Environment"
)

type resourceEnvironment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEnvironment) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_environment"
}

func (r *resourceEnvironment) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"engine_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": framework.IDAttribute(),
			// "high_availability_config": schema.ListAttribute{
			// 	CustomType:  fwtypes.NewListNestedObjectTypeOf[highAvailabilityConfig](ctx),
			// 	ElementType: fwtypes.NewObjectTypeOf[highAvailabilityConfig](ctx),
			// 	Optional:    true,
			// 	Computed:    true,
			// 	PlanModifiers: []planmodifier.List{
			// 		listplanmodifier.UseStateForUnknown(),
			// 	},
			// },
			"id": framework.IDAttribute(),
			"instance_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preferred_mainttainence_window": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"publicly_accessible": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"high_availability_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[highAvailabilityConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"desired_capacity": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},

			"storage_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[storageConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"efs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[efs](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"file_system_id": schema.StringAttribute{
										Required: true,
									},
									"mount_point": schema.StringAttribute{
										Required: true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"fsx": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[fsx](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"file_system_id": schema.StringAttribute{
										Required: true,
									},
									"mount_point": schema.StringAttribute{
										Required: true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
			},
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *resourceEnvironment) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().M2Client(ctx)
	var data resourceEnvironmentData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &m2.CreateEnvironmentInput{}

	response.Diagnostics.Append(flex.Expand(ctx, data, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	input.EngineType = awstypes.EngineType(*flex.StringFromFramework(ctx, data.EngineType))
	input.InstanceType = flex.StringFromFramework(ctx, data.InstanceType)
	input.Name = flex.StringFromFramework(ctx, data.Name)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateEnvironment(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameEnvironment, data.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := data
	state.ID = flex.StringToFramework(ctx, output.EnvironmentId)

	envARN := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "m2",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("env/%s", *output.EnvironmentId),
	}.String()

	state.ARN = types.StringValue(envARN)
	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	out, err := waitEnvironmentAvailable(ctx, conn, state.ID.ValueString(), createTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForCreation, ResNameEnvironment, data.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read implements resource.ResourceWithConfigure.
func (r *resourceEnvironment) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

	conn := r.Meta().M2Client(ctx)
	var data resourceEnvironmentData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	out, err := findEnvByID(ctx, conn, data.EnvironmentId.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameEnvironment, data.EnvironmentId.ValueString(), err),
			err.Error(),
		)
		return
	}

	envARN := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "m2",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("env/%s", *out.EnvironmentId),
	}.String()

	data.ARN = types.StringValue(envARN)

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.EnvironmentId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)

}

// Delete implements resource.ResourceWithConfigure.
func (r *resourceEnvironment) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().M2Client(ctx)
	var state resourceEnvironmentData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting M2 Environment", map[string]interface{}{
		"id": state.EnvironmentId.ValueString(),
	})

	input := &m2.DeleteEnvironmentInput{
		EnvironmentId: flex.StringFromFramework(ctx, state.EnvironmentId),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteEnvironment(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionDeleting, ResNameEnvironment, state.EnvironmentId.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitEnvironmentDeleted(ctx, conn, state.EnvironmentId.ValueString(), deleteTimeout)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForDeletion, ResNameEnvironment, state.EnvironmentId.ValueString(), err),
			err.Error(),
		)
		return
	}

}

// Update implements resource.ResourceWithConfigure.
func (r *resourceEnvironment) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().M2Client(ctx)
	var state, plan resourceEnvironmentData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if environmentHasChanges(ctx, plan, state) {
		input := &m2.UpdateEnvironmentInput{}
		response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

		if response.Diagnostics.HasError() {
			return
		}

		input.EnvironmentId = flex.StringFromFramework(ctx, state.ID)

		_, err := conn.UpdateEnvironment(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameEnvironment, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitEnvironmentAvailable(ctx, conn, state.ID.ValueString(), updateTimeout)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForUpdate, ResNameEnvironment, plan.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &plan)...)

	}

	// out, err := FindEnvByID(ctx, conn, state.ID.ValueString())

	// if err != nil {
	// 	response.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameEnvironment, state.ID.ValueString(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)

	// if response.Diagnostics.HasError() {
	// 	return
	// }

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)

}
func (r *resourceEnvironment) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("environment_id"), request, response)
}

func (r *resourceEnvironment) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceEnvironmentData struct {
	ARN                        types.String                                            `tfsdk:"arn"`
	ClientToken                types.String                                            `tfsdk:"client_token"`
	Description                types.String                                            `tfsdk:"description"`
	EngineType                 types.String                                            `tfsdk:"engine_type"`
	EngineVersion              types.String                                            `tfsdk:"engine_version"`
	EnvironmentId              types.String                                            `tfsdk:"environment_id"`
	HighAvailabilityConfig     fwtypes.ListNestedObjectValueOf[highAvailabilityConfig] `tfsdk:"high_availability_config"`
	ID                         types.String                                            `tfsdk:"id"`
	InstanceType               types.String                                            `tfsdk:"instance_type"`
	KmsKeyId                   types.String                                            `tfsdk:"kms_key_id"`
	Name                       types.String                                            `tfsdk:"name"`
	PreferredMaintenanceWindow types.String                                            `tfsdk:"preferred_mainttainence_window"`
	PubliclyAccessible         types.Bool                                              `tfsdk:"publicly_accessible"`
	SecurityGroupIds           types.Set                                               `tfsdk:"security_group_ids"`
	StorageConfiguration       fwtypes.ListNestedObjectValueOf[storageConfiguration]   `tfsdk:"storage_configuration"`
	SubnetIds                  types.Set                                               `tfsdk:"subnet_ids"`
	Tags                       types.Map                                               `tfsdk:"tags"`
	TagsAll                    types.Map                                               `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                          `tfsdk:"timeouts"`
}

type storageConfiguration struct {
	EFS fwtypes.ListNestedObjectValueOf[efs] `tfsdk:"efs"`
	FSX fwtypes.ListNestedObjectValueOf[fsx] `tfsdk:"fsx"`
}

type efs struct {
	FileSystemId types.String `tfsdk:"file_system_id"`
	MountPoint   types.String `tfsdk:"mount_point"`
}

type fsx struct {
	FileSystemId types.String `tfsdk:"file_system_id"`
	MountPoint   types.String `tfsdk:"mount_point"`
}

type highAvailabilityConfig struct {
	DesiredCapacity types.Int64 `tfsdk:"desired_capacity"`
}

func environmentHasChanges(_ context.Context, plan, state resourceEnvironmentData) bool {
	return !plan.EngineType.Equal(state.EngineType) ||
		!plan.Description.Equal(state.Description) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.EngineVersion.Equal(state.EngineVersion) ||
		!plan.HighAvailabilityConfig.Equal(state.HighAvailabilityConfig) ||
		!plan.InstanceType.Equal(state.InstanceType) ||
		!plan.KmsKeyId.Equal(state.InstanceType) ||
		!plan.Name.Equal(state.Name) ||
		!plan.PreferredMaintenanceWindow.Equal(state.PreferredMaintenanceWindow) ||
		!plan.PubliclyAccessible.Equal(state.PubliclyAccessible) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.StorageConfiguration.Equal(state.StorageConfiguration) ||
		!plan.SubnetIds.Equal(state.SubnetIds)

}
