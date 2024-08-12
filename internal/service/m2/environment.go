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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_m2_environment", name="Environment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/m2;m2.GetEnvironmentOutput")
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

func (*environmentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_environment"
}

func (r *environmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apply_changes_during_maintenance_window": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
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
			names.AttrEngineVersion: schema.StringAttribute{
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
			"force_update": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrInstanceType: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^\S{1,20}$`), ""),
				},
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"load_balancer_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{1,59}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPreferredMaintenanceWindow: schema.StringAttribute{
				CustomType: fwtypes.OnceAWeekWindowType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
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
			names.AttrSubnetIDs: schema.SetAttribute{
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[efsStorageConfigurationModel](ctx),
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
									names.AttrFileSystemID: schema.StringAttribute{
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[fsxStorageConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrFileSystemID: schema.StringAttribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
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

	name := data.Name.ValueString()
	input := m2.CreateEnvironmentInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateEnvironment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Mainframe Modernization Environment (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.EnvironmentID = fwflex.StringToFramework(ctx, output.EnvironmentId)
	data.setID()

	env, err := waitEnvironmentCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
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
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().M2Client(ctx)

	output, err := findEnvironmentByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Mainframe Modernization Environment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *environmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	if !new.EngineVersion.Equal(old.EngineVersion) ||
		!new.HighAvailabilityConfig.Equal(old.HighAvailabilityConfig) ||
		!new.InstanceType.Equal(old.InstanceType) ||
		!new.PreferredMaintenanceWindow.Equal(old.PreferredMaintenanceWindow) {
		input := &m2.UpdateEnvironmentInput{
			EnvironmentId: fwflex.StringFromFramework(ctx, new.ID),
		}

		if !new.ForceUpdate.IsNull() {
			input.ForceUpdate = new.ForceUpdate.ValueBool()
		}

		// https://docs.aws.amazon.com/m2/latest/APIReference/API_UpdateEnvironment.html#m2-UpdateEnvironment-request-applyDuringMaintenanceWindow.
		// "Currently, AWS Mainframe Modernization accepts the engineVersion parameter only if applyDuringMaintenanceWindow is true. If any parameter other than engineVersion is provided in UpdateEnvironmentRequest, it will fail if applyDuringMaintenanceWindow is set to true."
		if new.ApplyDuringMaintenanceWindow.ValueBool() && !new.EngineVersion.Equal(old.EngineVersion) {
			input.ApplyDuringMaintenanceWindow = true
			input.EngineVersion = fwflex.StringFromFramework(ctx, new.EngineVersion)
		} else {
			if !new.EngineVersion.Equal(old.EngineVersion) {
				input.EngineVersion = fwflex.StringFromFramework(ctx, new.EngineVersion)
			}
			if !new.HighAvailabilityConfig.Equal(old.HighAvailabilityConfig) {
				highAvailabilityConfigData, diags := new.HighAvailabilityConfig.ToPtr(ctx)
				response.Diagnostics.Append(diags...)
				if response.Diagnostics.HasError() {
					return
				}

				input.DesiredCapacity = fwflex.Int32FromFramework(ctx, highAvailabilityConfigData.DesiredCapacity)
			}
			if !new.InstanceType.Equal(old.InstanceType) {
				input.InstanceType = fwflex.StringFromFramework(ctx, new.InstanceType)
			}
			if !new.PreferredMaintenanceWindow.Equal(old.PreferredMaintenanceWindow) {
				input.PreferredMaintenanceWindow = fwflex.StringFromFramework(ctx, new.PreferredMaintenanceWindow)
			}
		}
		_, err := conn.UpdateEnvironment(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Mainframe Modernization Environment (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitEnvironmentUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Environment (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	_, err := conn.DeleteEnvironment(ctx, &m2.DeleteEnvironmentInput{
		EnvironmentId: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Mainframe Modernization Environment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Environment (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *environmentResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
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
		Pending:    enum.Slice(awstypes.EnvironmentLifecycleAvailable, awstypes.EnvironmentLifecycleCreating, awstypes.EnvironmentLifecycleDeleting),
		Target:     []string{},
		Refresh:    statusEnvironment(ctx, conn, id),
		Timeout:    timeout,
		Delay:      4 * time.Minute,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type environmentResourceModel struct {
	ApplyDuringMaintenanceWindow types.Bool                                                   `tfsdk:"apply_changes_during_maintenance_window"`
	Description                  types.String                                                 `tfsdk:"description"`
	EngineType                   fwtypes.StringEnum[awstypes.EngineType]                      `tfsdk:"engine_type"`
	EngineVersion                types.String                                                 `tfsdk:"engine_version"`
	EnvironmentARN               types.String                                                 `tfsdk:"arn"`
	EnvironmentID                types.String                                                 `tfsdk:"environment_id"`
	ForceUpdate                  types.Bool                                                   `tfsdk:"force_update"`
	HighAvailabilityConfig       fwtypes.ListNestedObjectValueOf[highAvailabilityConfigModel] `tfsdk:"high_availability_config"`
	ID                           types.String                                                 `tfsdk:"id"`
	InstanceType                 types.String                                                 `tfsdk:"instance_type"`
	KmsKeyID                     fwtypes.ARN                                                  `tfsdk:"kms_key_id"`
	LoadBalancerArn              types.String                                                 `tfsdk:"load_balancer_arn"`
	Name                         types.String                                                 `tfsdk:"name"`
	PreferredMaintenanceWindow   fwtypes.OnceAWeekWindow                                      `tfsdk:"preferred_maintenance_window"`
	PubliclyAccessible           types.Bool                                                   `tfsdk:"publicly_accessible"`
	SecurityGroupIDs             fwtypes.SetValueOf[types.String]                             `tfsdk:"security_group_ids"`
	StorageConfigurations        fwtypes.ListNestedObjectValueOf[storageConfigurationModel]   `tfsdk:"storage_configuration"`
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
	EFS fwtypes.ListNestedObjectValueOf[efsStorageConfigurationModel] `tfsdk:"efs"`
	FSX fwtypes.ListNestedObjectValueOf[fsxStorageConfigurationModel] `tfsdk:"fsx"`
}

var (
	_ fwflex.Expander  = storageConfigurationModel{}
	_ fwflex.Flattener = &storageConfigurationModel{}
)

func (m storageConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.EFS.IsNull():
		efsStorageConfigurationData, d := m.EFS.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.StorageConfigurationMemberEfs
		diags.Append(fwflex.Expand(ctx, efsStorageConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.FSX.IsNull():
		fsxStorageConfigurationData, d := m.FSX.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.StorageConfigurationMemberFsx
		diags.Append(fwflex.Expand(ctx, fsxStorageConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *storageConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.StorageConfigurationMemberEfs:
		var model efsStorageConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.EFS = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	case awstypes.StorageConfigurationMemberFsx:
		var model fsxStorageConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.FSX = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	default:
		return diags
	}
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
