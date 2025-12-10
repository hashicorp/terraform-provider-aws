// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_odb_cloud_exadata_infrastructure", name="Cloud Exadata Infrastructure")
// @Tags(identifierAttribute="arn")
func newResourceCloudExadataInfrastructure(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCloudExadataInfrastructure{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameCloudExadataInfrastructure = "Cloud Exadata Infrastructure"
)

type resourceCloudExadataInfrastructure struct {
	framework.ResourceWithModel[cloudExadataInfrastructureResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceCloudExadataInfrastructure) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activated_storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of storage servers requested for the Exadata infrastructure",
			},
			"additional_storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: " The number of storage servers requested for the Exadata infrastructure",
			},
			"database_server_type": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The database server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation",
			},
			"storage_server_type": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The storage server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation",
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"available_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of available storage, in gigabytes (GB), for the Exadata infrastructure",
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the Availability Zone (AZ) where the Exadata infrastructure is located. Changing this will force terraform to create new resource",
			},
			"availability_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: " The AZ ID of the AZ where the Exadata infrastructure is located. Changing this will force terraform to create new resource",
			},
			"compute_count": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Description: " The number of compute instances that the Exadata infrastructure is located",
			},
			"cpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores that are allocated to the Exadata infrastructure",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The size of the Exadata infrastructure's data disk group, in terabytes (TB)",
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The size of the Exadata infrastructure's local node storage, in gigabytes (GB)",
			},
			"db_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The software version of the database servers (dom0) in the Exadata infrastructure",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The user-friendly name for the Exadata infrastructure. Changing this will force terraform to create a new resource",
			},
			"last_maintenance_run_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud Identifier (OCID) of the last maintenance run for the Exadata infrastructure",
			},
			"max_cpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores available on the Exadata infrastructure",
			},
			"max_data_storage_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total amount of data disk group storage, in terabytes (TB), that's available on the Exadata infrastructure",
			},
			"max_db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of local node storage, in gigabytes (GB), that's available on the Exadata infrastructure",
			},
			"max_memory_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of memory in gigabytes (GB) available on the Exadata infrastructure",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of memory, in gigabytes (GB), that's allocated on the Exadata infrastructure",
			},
			"monthly_db_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The monthly software version of the database servers in the Exadata infrastructure",
			},
			"monthly_storage_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The monthly software version of the storage servers installed on the Exadata infrastructure",
			},
			"next_maintenance_run_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the next maintenance run for the Exadata infrastructure",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the Exadata infrastructure",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor for the Exadata infrastructure",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "The HTTPS link to the Exadata infrastructure in OCI",
			},
			"percent_progress": schema.Float64Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the Exadata infrastructure, expressed as a percentage",
			},
			"shape": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The model name of the Exadata infrastructure. Changing this will force terraform to create new resource",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The current status of the Exadata infrastructure",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the Exadata infrastructure",
			},
			"storage_count": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "TThe number of storage servers that are activated for the Exadata infrastructure",
			},
			"storage_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The software version of the storage servers on the Exadata infrastructure.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"total_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of storage, in gigabytes (GB), on the Exadata infrastructure.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The time when the Exadata infrastructure was created.",
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
				Description: "The OCI model compute model used when you create or clone an\n " +
					" instance: ECPU or OCPU. An ECPU is an abstracted measure of\n " +
					"compute resources. ECPUs are based on the number of cores\n " +
					"elastically allocated from a pool of compute and storage servers.\n " +
					" An OCPU is a legacy physical measure of compute resources. OCPUs\n " +
					"are based on the physical core of a processor with\n " +
					" hyper-threading enabled.",
			},
			"customer_contacts_to_send_to_oci": schema.SetAttribute{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[customerContactExaInfraResourceModel](ctx),
				Optional:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The email addresses of contacts to receive notification from Oracle about maintenance updates for the Exadata infrastructure. Changing this will force terraform to create new resource",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"maintenance_window": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudExadataInfraMaintenanceWindowResourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				Description: " The scheduling details for the maintenance window. Patching and system updates take place during the maintenance window ",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"custom_action_timeout_in_mins": schema.Int32Attribute{
							Required: true,
						},
						"days_of_week": schema.SetAttribute{
							ElementType: fwtypes.NewObjectTypeOf[dayOfWeekExaInfraMaintenanceWindowResourceModel](ctx),
							Optional:    true,
							Computed:    true,
						},
						"hours_of_day": schema.SetAttribute{
							ElementType: types.Int32Type,
							Optional:    true,
							Computed:    true,
						},
						"is_custom_action_timeout_enabled": schema.BoolAttribute{
							Required: true,
						},
						"lead_time_in_weeks": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						"months": schema.SetAttribute{
							ElementType: fwtypes.NewObjectTypeOf[monthExaInfraMaintenanceWindowResourceModel](ctx),
							Optional:    true,
							Computed:    true,
						},
						"patching_mode": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[odbtypes.PatchingModeType](),
						},
						"preference": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[odbtypes.PreferenceType](),
						},
						"weeks_of_month": schema.SetAttribute{
							ElementType: types.Int32Type,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceCloudExadataInfrastructure) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan cloudExadataInfrastructureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateCloudExadataInfrastructureInput{
		Tags: getTagsIn(ctx),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CloudExadataInfrastructureId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdExaInfra, err := waitCloudExadataInfrastructureCreated(ctx, conn, aws.ToString(out.CloudExadataInfrastructureId), createTimeout)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(out.CloudExadataInfrastructureId))...)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, createdExaInfra, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCloudExadataInfrastructure) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findExadataInfraResourceByID(ctx, conn, state.CloudExadataInfrastructureId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
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

func (r *resourceCloudExadataInfrastructure) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().ODBClient(ctx)

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if diff.HasChanges() {
		updatedMW := odb.UpdateCloudExadataInfrastructureInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &updatedMW)...)

		out, err := conn.UpdateCloudExadataInfrastructure(ctx, &updatedMW)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
				err.Error(),
			)
			return
		}
	}
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedExaInfra, err := waitCloudExadataInfrastructureUpdated(ctx, conn, state.CloudExadataInfrastructureId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, updatedExaInfra, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCloudExadataInfrastructure) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: state.CloudExadataInfrastructureId.ValueStringPointer(),
	}

	_, err := conn.DeleteCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCloudExadataInfrastructureDeleted(ctx, conn, state.CloudExadataInfrastructureId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}
}

func waitCloudExadataInfrastructureCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:      enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:       enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:      statusCloudExadataInfrastructure(ctx, conn, id),
		PollInterval: 1 * time.Minute,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}
	return nil, err
}

func waitCloudExadataInfrastructureUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:      enum.Slice(odbtypes.ResourceStatusUpdating),
		Target:       enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:      statusCloudExadataInfrastructure(ctx, conn, id),
		PollInterval: 1 * time.Minute,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}

	return nil, err
}

func waitCloudExadataInfrastructureDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusCloudExadataInfrastructure(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}

	return nil, err
}

func statusCloudExadataInfrastructure(ctx context.Context, conn *odb.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findExadataInfraResourceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findExadataInfraResourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudExadataInfrastructure, error) {
	input := odb.GetCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: aws.String(id),
	}

	out, err := conn.GetCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.CloudExadataInfrastructure == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.CloudExadataInfrastructure, nil
}

type cloudExadataInfrastructureResourceModel struct {
	framework.WithRegionModel
	ActivatedStorageCount         types.Int32                                                                      `tfsdk:"activated_storage_count"`
	AdditionalStorageCount        types.Int32                                                                      `tfsdk:"additional_storage_count"`
	DatabaseServerType            types.String                                                                     `tfsdk:"database_server_type" `
	StorageServerType             types.String                                                                     `tfsdk:"storage_server_type" `
	AvailabilityZone              types.String                                                                     `tfsdk:"availability_zone"`
	AvailabilityZoneId            types.String                                                                     `tfsdk:"availability_zone_id"`
	AvailableStorageSizeInGBs     types.Int32                                                                      `tfsdk:"available_storage_size_in_gbs"`
	CloudExadataInfrastructureArn types.String                                                                     `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String                                                                     `tfsdk:"id"`
	ComputeCount                  types.Int32                                                                      `tfsdk:"compute_count"`
	CpuCount                      types.Int32                                                                      `tfsdk:"cpu_count"`
	CustomerContactsToSendToOCI   fwtypes.SetNestedObjectValueOf[customerContactExaInfraResourceModel]             `tfsdk:"customer_contacts_to_send_to_oci"`
	DataStorageSizeInTBs          types.Float64                                                                    `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                                      `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerVersion               types.String                                                                     `tfsdk:"db_server_version"`
	DisplayName                   types.String                                                                     `tfsdk:"display_name"`
	LastMaintenanceRunId          types.String                                                                     `tfsdk:"last_maintenance_run_id"`
	MaxCpuCount                   types.Int32                                                                      `tfsdk:"max_cpu_count"`
	MaxDataStorageInTBs           types.Float64                                                                    `tfsdk:"max_data_storage_in_tbs"`
	MaxDbNodeStorageSizeInGBs     types.Int32                                                                      `tfsdk:"max_db_node_storage_size_in_gbs"`
	MaxMemoryInGBs                types.Int32                                                                      `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs               types.Int32                                                                      `tfsdk:"memory_size_in_gbs"`
	MonthlyDbServerVersion        types.String                                                                     `tfsdk:"monthly_db_server_version"`
	MonthlyStorageServerVersion   types.String                                                                     `tfsdk:"monthly_storage_server_version"`
	NextMaintenanceRunId          types.String                                                                     `tfsdk:"next_maintenance_run_id"`
	Ocid                          types.String                                                                     `tfsdk:"ocid"`
	OciResourceAnchorName         types.String                                                                     `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                                     `tfsdk:"oci_url"`
	PercentProgress               types.Float64                                                                    `tfsdk:"percent_progress"`
	Shape                         types.String                                                                     `tfsdk:"shape"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                                      `tfsdk:"status"`
	StatusReason                  types.String                                                                     `tfsdk:"status_reason"`
	StorageCount                  types.Int32                                                                      `tfsdk:"storage_count"`
	StorageServerVersion          types.String                                                                     `tfsdk:"storage_server_version"`
	TotalStorageSizeInGBs         types.Int32                                                                      `tfsdk:"total_storage_size_in_gbs"`
	Timeouts                      timeouts.Value                                                                   `tfsdk:"timeouts"`
	CreatedAt                     timetypes.RFC3339                                                                `tfsdk:"created_at" `
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                                        `tfsdk:"compute_model"`
	Tags                          tftags.Map                                                                       `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                                       `tfsdk:"tags_all"`
	MaintenanceWindow             fwtypes.ListNestedObjectValueOf[cloudExadataInfraMaintenanceWindowResourceModel] `tfsdk:"maintenance_window"`
}

type cloudExadataInfraMaintenanceWindowResourceModel struct {
	CustomActionTimeoutInMins    types.Int32                                                                     `tfsdk:"custom_action_timeout_in_mins"`
	DaysOfWeek                   fwtypes.SetNestedObjectValueOf[dayOfWeekExaInfraMaintenanceWindowResourceModel] `tfsdk:"days_of_week" `
	HoursOfDay                   fwtypes.SetValueOf[types.Int64]                                                 `tfsdk:"hours_of_day"`
	IsCustomActionTimeoutEnabled types.Bool                                                                      `tfsdk:"is_custom_action_timeout_enabled"`
	LeadTimeInWeeks              types.Int32                                                                     `tfsdk:"lead_time_in_weeks"`
	Months                       fwtypes.SetNestedObjectValueOf[monthExaInfraMaintenanceWindowResourceModel]     `tfsdk:"months" `
	PatchingMode                 fwtypes.StringEnum[odbtypes.PatchingModeType]                                   `tfsdk:"patching_mode"`
	Preference                   fwtypes.StringEnum[odbtypes.PreferenceType]                                     `tfsdk:"preference"`
	WeeksOfMonth                 fwtypes.SetValueOf[types.Int64]                                                 `tfsdk:"weeks_of_month"`
}

type dayOfWeekExaInfraMaintenanceWindowResourceModel struct {
	Name fwtypes.StringEnum[odbtypes.DayOfWeekName] `tfsdk:"name"`
}

type monthExaInfraMaintenanceWindowResourceModel struct {
	Name fwtypes.StringEnum[odbtypes.MonthName] `tfsdk:"name"`
}

type customerContactExaInfraResourceModel struct {
	Email types.String `tfsdk:"email"`
}
