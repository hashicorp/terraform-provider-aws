// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Environment")
// @Tags(identifierAttribute="arn")
func newEnvironmentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type environmentResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *environmentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_environment"
}

func (r *environmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apply_changes_during_maintenance_window": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(500),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EngineType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^\S{1,10}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": framework.IDAttribute(),
			"force_update": schema.BoolAttribute{ // TODO ????
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^\S{1,20}$`), ""),
				},
			},
			"kms_key_id": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"load_balancer_arn": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{1,59}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preferred_maintenance_window": schema.StringAttribute{ // TODO Custom type?
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
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(2),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"high_availability_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[highAvailabilityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"desired_capacity": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
					},
				},
			},
			"storage_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[storageConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"efs": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("efs"),
									path.MatchRelative().AtParent().AtName("fsx"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"file_system_id": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"mount_point": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"fsx": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"file_system_id": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"mount_point": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
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

func (r *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	// We can't use AutoFlEx for StorageConfigurations as it's a union.
	var storageConfigurations []awstypes.StorageConfiguration
	if !data.StorageConfiguration.IsNull() {
		var storageConfigurationModels []storageConfigurationModel
		response.Diagnostics.Append(data.StorageConfiguration.ElementsAs(ctx, &storageConfigurationModels, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		var d diag.Diagnostics
		storageConfigurations, d = expandStorageConfigurations(ctx, storageConfigurationModels)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	name := data.Name.ValueString()
	input := &m2.CreateEnvironmentInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)
	input.StorageConfigurations = storageConfigurations

	output, err := conn.CreateEnvironment(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Mainframe Modernization Environment (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.EnvironmentID = flex.StringToFramework(ctx, output.EnvironmentId)
	data.setID()

	env, err := waitEnvironmentCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Environment (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, env, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().M2Client(ctx)

	var state environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameEnvironment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *environmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().M2Client(ctx)

	var plan, state environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in, updateRequired := r.updateEnvironmentInput(ctx, plan, state, response)
	if !updateRequired {
		return
	}

	out, err := conn.UpdateEnvironment(ctx, in)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameEnvironment, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.EnvironmentId == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameEnvironment, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.EnvironmentId)

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	env, err := waitEnvironmentUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForUpdate, ResNameEnvironment, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, env, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *environmentResource) updateEnvironmentInput(ctx context.Context, plan, state environmentResourceModel, resp *resource.UpdateResponse) (*m2.UpdateEnvironmentInput, bool) {
	in := &m2.UpdateEnvironmentInput{
		EnvironmentId: flex.StringFromFramework(ctx, plan.ID),
	}

	if r.hasChangesForMaintenance(plan, state) {
		in.ApplyDuringMaintenanceWindow = true
		in.EngineVersion = flex.StringFromFramework(ctx, plan.EngineVersion)
	} else if r.hasChanges(plan, state) {
		if !plan.EngineVersion.Equal(state.EngineVersion) {
			in.EngineVersion = flex.StringFromFramework(ctx, plan.EngineVersion)
		}
		if !plan.InstanceType.Equal(state.InstanceType) {
			in.InstanceType = flex.StringFromFramework(ctx, plan.InstanceType)
		}
		if !plan.PreferredMaintenanceWindow.Equal(state.PreferredMaintenanceWindow) {
			in.PreferredMaintenanceWindow = flex.StringFromFramework(ctx, plan.PreferredMaintenanceWindow)
		}

		if !plan.HighAvailabilityConfig.Equal(state.HighAvailabilityConfig) {
			v, d := plan.HighAvailabilityConfig.ToSlice(ctx)
			resp.Diagnostics.Append(d...)
			if len(v) > 0 {
				in.DesiredCapacity = flex.Int32FromFramework(ctx, v[0].DesiredCapacity)
			}
		}
	} else {
		return nil, false
	}

	if !plan.ForceUpdate.IsNull() {
		in.ForceUpdate = plan.ForceUpdate.ValueBool()
	}
	return in, true
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().M2Client(ctx)

	var state environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &m2.DeleteEnvironmentInput{
		EnvironmentId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteEnvironment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionDeleting, ResNameEnvironment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitEnvironmentDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForDeletion, ResNameEnvironment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findEnvironmentByID(ctx context.Context, conn *m2.Client, id string) (*m2.GetEnvironmentOutput, error) {
	input := &m2.GetEnvironmentInput{
		EnvironmentId: aws.String(id),
	}

	output, err := conn.GetEnvironment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
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

func statusEnvironment(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEnvironmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEnvironmentCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentLifecycleCreating),
		Target:  enum.Slice(awstypes.EnvironmentLifecycleAvailable),
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitEnvironmentUpdated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentLifecycleUpdating),
		Target:  enum.Slice(awstypes.EnvironmentLifecycleAvailable),
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentLifecycleAvailable, awstypes.EnvironmentLifecycleCreating, awstypes.EnvironmentLifecycleDeleting),
		Target:  []string{},
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func (rd *environmentResourceModel) refreshFromOutput(ctx context.Context, out *m2.GetEnvironmentOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, out, rd)...)
	rd.ARN = flex.StringToFramework(ctx, out.EnvironmentArn)
	rd.ID = flex.StringToFramework(ctx, out.EnvironmentId)
	storage, d := flattenStorageConfigurations(ctx, out.StorageConfigurations)
	diags.Append(d...)
	rd.StorageConfiguration = storage

	return diags
}

type environmentResourceModel struct {
	ARN                          types.String                                                 `tfsdk:"arn"`
	ApplyDuringMaintenanceWindow types.Bool                                                   `tfsdk:"apply_changes_during_maintenance_window"`
	Description                  types.String                                                 `tfsdk:"description"`
	EngineType                   fwtypes.StringEnum[awstypes.EngineType]                      `tfsdk:"engine_type"`
	EngineVersion                types.String                                                 `tfsdk:"engine_version"`
	EnvironmentID                types.String                                                 `tfsdk:"environment_id"`
	ForceUpdate                  types.Bool                                                   `tfsdk:"force_update"`
	HighAvailabilityConfig       fwtypes.ListNestedObjectValueOf[highAvailabilityConfigModel] `tfsdk:"high_availability_config"`
	ID                           types.String                                                 `tfsdk:"id"`
	InstanceType                 types.String                                                 `tfsdk:"instance_type"`
	KmsKeyID                     fwtypes.ARN                                                  `tfsdk:"kms_key_id"`
	LoadBalancerArn              types.String                                                 `tfsdk:"load_balancer_arn"`
	Name                         types.String                                                 `tfsdk:"name"`
	PreferredMaintenanceWindow   types.String                                                 `tfsdk:"preferred_maintenance_window"`
	PubliclyAccessible           types.Bool                                                   `tfsdk:"publicly_accessible"`
	SecurityGroupIDs             fwtypes.SetValueOf[types.String]                             `tfsdk:"security_group_ids"`
	StorageConfiguration         fwtypes.ListNestedObjectValueOf[storageConfigurationModel]   `tfsdk:"storage_configuration"`
	SubnetIDs                    fwtypes.SetValueOf[types.String]                             `tfsdk:"subnet_ids"`
	Tags                         types.Map                                                    `tfsdk:"tags"`
	TagsAll                      types.Map                                                    `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                               `tfsdk:"timeouts"`
}

func (model *environmentResourceModel) InitFromID() error {
	model.EnvironmentID = model.ID

	return nil
}

func (model *environmentResourceModel) setID() {
	model.ID = model.EnvironmentID
}

type storageConfigurationModel struct {
	EFS types.List `tfsdk:"efs"`
	FSX types.List `tfsdk:"fsx"`
}

type efsStorageConfigurationModel struct {
	FileSystemID types.String `tfsdk:"file_system_id"`
	MountPoint   types.String `tfsdk:"mount_point"`
}

type fsxStorageConfigurationModel struct {
	FileSystemID types.String `tfsdk:"file_system_id"`
	MountPoint   types.String `tfsdk:"mount_point"`
}

type highAvailabilityConfigModel struct {
	DesiredCapacity types.Int64 `tfsdk:"desired_capacity"`
}

var (
	storageDataAttrTypes = map[string]attr.Type{
		"efs": types.ListType{ElemType: mountObjectType},
		"fsx": types.ListType{ElemType: mountObjectType},
	}

	mountObjectType = types.ObjectType{AttrTypes: mountAttrTypes}

	mountAttrTypes = map[string]attr.Type{
		"file_system_id": types.StringType,
		"mount_point":    types.StringType,
	}
)

func (r *environmentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func expandStorageConfigurations(ctx context.Context, storageConfigurationModels []storageConfigurationModel) ([]awstypes.StorageConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	storageConfigurations := []awstypes.StorageConfiguration{}

	for _, item := range storageConfigurationModels {
		if !item.EFS.IsNull() && (len(item.EFS.Elements()) > 0) {
			var efsStorageConfigurationModels []efsStorageConfigurationModel
			diags.Append(item.EFS.ElementsAs(ctx, &efsStorageConfigurationModels, false)...)
			v := expandStorageConfigurationMemberEFS(ctx, efsStorageConfigurationModels)
			storageConfigurations = append(storageConfigurations, v)
		}
		if !item.FSX.IsNull() && (len(item.FSX.Elements()) > 0) {
			var fsxStorageConfigurationModels []fsxStorageConfigurationModel
			diags.Append(item.FSX.ElementsAs(ctx, &fsxStorageConfigurationModels, false)...)
			v := expandStorageConfigurationMemberFSX(ctx, fsxStorageConfigurationModels)
			storageConfigurations = append(storageConfigurations, v)
		}
	}

	return storageConfigurations, diags
}

func expandStorageConfigurationMemberEFS(ctx context.Context, efsStorageConfigurationModels []efsStorageConfigurationModel) *awstypes.StorageConfigurationMemberEfs {
	if len(efsStorageConfigurationModels) == 0 {
		return nil
	}

	return &awstypes.StorageConfigurationMemberEfs{
		Value: awstypes.EfsStorageConfiguration{
			FileSystemId: flex.StringFromFramework(ctx, efsStorageConfigurationModels[0].FileSystemID),
			MountPoint:   flex.StringFromFramework(ctx, efsStorageConfigurationModels[0].MountPoint),
		},
	}
}

func expandStorageConfigurationMemberFSX(ctx context.Context, fsxStorageConfigurationModels []fsxStorageConfigurationModel) *awstypes.StorageConfigurationMemberFsx {
	if len(fsxStorageConfigurationModels) == 0 {
		return nil
	}

	return &awstypes.StorageConfigurationMemberFsx{
		Value: awstypes.FsxStorageConfiguration{
			FileSystemId: flex.StringFromFramework(ctx, fsxStorageConfigurationModels[0].FileSystemID),
			MountPoint:   flex.StringFromFramework(ctx, fsxStorageConfigurationModels[0].MountPoint),
		},
	}
}

func flattenStorageConfigurations(ctx context.Context, apiObject []awstypes.StorageConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: storageDataAttrTypes}

	elems := []attr.Value{}

	for _, config := range apiObject {
		switch v := config.(type) {
		case *awstypes.StorageConfigurationMemberEfs:
			mountPoint, d := flattenMountPoint(ctx, v.Value.FileSystemId, v.Value.MountPoint, "efs")
			elems = append(elems, mountPoint)
			diags.Append(d...)

		case *awstypes.StorageConfigurationMemberFsx:
			mountPoint, d := flattenMountPoint(ctx, v.Value.FileSystemId, v.Value.MountPoint, "fsx")
			elems = append(elems, mountPoint)
			diags.Append(d...)
		}
	}
	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenMountPoint(ctx context.Context, fileSystemId, mountPoint *string, mountType string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	obj := map[string]attr.Value{
		"file_system_id": flex.StringToFramework(ctx, fileSystemId),
		"mount_point":    flex.StringToFramework(ctx, mountPoint),
	}

	mountValue, d := types.ObjectValue(mountAttrTypes, obj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	mountList := []attr.Value{
		mountValue,
	}

	mountListValue, d := types.ListValue(mountObjectType, mountList)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	configMap := map[string]attr.Value{
		mountType: mountListValue,
	}

	for k := range storageDataAttrTypes {
		if k != mountType {
			configMap[k] = types.ListNull(mountObjectType)
		}
	}

	configValue, d := types.ObjectValue(storageDataAttrTypes, configMap)
	diags.Append(d...)

	return configValue, diags
}

func (r *environmentResource) hasChanges(plan, state environmentResourceModel) bool {
	return !plan.HighAvailabilityConfig.Equal(state.HighAvailabilityConfig) ||
		!plan.EngineVersion.Equal(state.EngineVersion) ||
		!plan.InstanceType.Equal(state.EngineType) ||
		!plan.PreferredMaintenanceWindow.Equal(state.PreferredMaintenanceWindow)
}

func (r *environmentResource) hasChangesForMaintenance(plan, state environmentResourceModel) bool {
	return plan.ApplyDuringMaintenanceWindow.ValueBool() && !plan.EngineVersion.Equal(state.EngineVersion)
}
