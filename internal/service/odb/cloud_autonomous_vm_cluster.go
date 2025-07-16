// Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_odb_cloud_autonomous_vm_cluster", name="Cloud Autonomous Vm Cluster")
// @Tags(identifierAttribute="arn")
func newResourceCloudAutonomousVmCluster(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCloudAutonomousVmCluster{}
	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameCloudAutonomousVmCluster = "Cloud Autonomous Vm Cluster"
	NotAvailableValues              = "NOT_AVAILABLE"
)

var ResourceCloudAutonomousVMCluster = newResourceCloudAutonomousVmCluster

type resourceCloudAutonomousVmCluster struct {
	framework.ResourceWithModel[cloudAutonomousVmClusterResourceModel]
	framework.WithTimeouts
}

func (r *resourceCloudAutonomousVmCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	status := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	licenseModel := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	computeModel := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	stringLengthBetween1And255Validator := []validator.String{
		stringvalidator.LengthBetween(1, 255),
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Exadata infrastructure id. Changing this will force terraform to create new resource.",
			},
			"autonomous_data_storage_percentage": schema.Float32Attribute{
				Computed:    true,
				Description: "The progress of the current operation on the Autonomous VM cluster, as a percentage.",
			},
			"autonomous_data_storage_size_in_tbs": schema.Float64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
				Description: "The data storage size allocated for Autonomous Databases in the Autonomous VM cluster, in TB. Changing this will force terraform to create new resource.",
			},
			"available_autonomous_data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.",
			},
			"available_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that you can create with the currently available storage.",
			},
			"available_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores available for allocation to Autonomous Databases",
			},
			"compute_model": schema.StringAttribute{
				CustomType:  computeModel,
				Computed:    true,
				Description: "The compute model of the Autonomous VM cluster: ECPU or OCPU.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores in the Autonomous VM cluster.",
			},
			"cpu_core_count_per_node": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The number of CPU cores enabled per node in the Autonomous VM cluster.",
			},
			"cpu_percentage": schema.Float32Attribute{
				Computed:    true,
				Description: "The percentage of total CPU cores currently in use in the Autonomous VM cluster.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the Autonomous VM cluster was created.",
			},
			"data_storage_size_in_gbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total data storage allocated to the Autonomous VM cluster, in GB.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total data storage allocated to the Autonomous VM cluster, in TB.",
			},
			"odb_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: " The local node storage allocated to the Autonomous VM cluster, in gigabytes (GB)",
			},
			"db_servers": schema.SetAttribute{
				Required:    true,
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Description: "The database servers in the Autonomous VM cluster. Changing this will force terraform to create new resource.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The description of the Autonomous VM cluster.",
			},
			"display_name": schema.StringAttribute{
				Required:   true,
				Validators: stringLengthBetween1And255Validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The display name of the Autonomous VM cluster. Changing this will force terraform to create new resource.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The domain name of the Autonomous VM cluster.",
			},
			"exadata_storage_in_tbs_lowest_scaled_value": schema.Float64Attribute{
				Computed:    true,
				Description: "The minimum value to which you can scale down the Exadata storage, in TB.",
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The hostname of the Autonomous VM cluster.",
			},
			"is_mtls_enabled_vm_cluster": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether mutual TLS (mTLS) authentication is enabled for the Autonomous VM cluster. Changing this will force terraform to create new resource. ",
			},
			"license_model": schema.StringAttribute{
				CustomType: licenseModel,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The license model for the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE . Changing this will force terraform to create new resource.",
			},
			"max_acds_lowest_scaled_value": schema.Int32Attribute{
				Computed:    true,
				Description: "The minimum value to which you can scale down the maximum number of Autonomous CDBs.",
			},
			"memory_per_oracle_compute_unit_in_gbs": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The amount of memory allocated per Oracle Compute Unit, in GB. Changing this will force terraform to create new resource.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of memory allocated to the Autonomous VM cluster, in gigabytes(GB).",
			},
			"node_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of database server nodes in the Autonomous VM cluster.",
			},
			"non_provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that can't be provisioned because of resource constraints.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor associated with this Autonomous VM cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for accessing the OCI console page for this Autonomous VM cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.",
			},
			"odb_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The unique identifier of the ODB network associated with this Autonomous VM Cluster. Changing this will force terraform to create new resource.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: `The progress of the current operation on the Autonomous VM cluster, as a percentage.`,
			},
			"provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.",
			},
			"provisioned_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.",
			},
			"provisioned_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPUs provisioned in the Autonomous VM cluster.",
			},
			"reclaimable_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.",
			},
			"reserved_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores reserved for system operations and redundancy.",
			},
			"scan_listener_port_non_tls": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The SCAN listener port for non-TLS (TCP) protocol. The default is 1521. Changing this will force terraform to create new resource.",
			},
			"scan_listener_port_tls": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The SCAN listener port for TLS (TCP) protocol. The default is 2484. Changing this will force terraform to create new resource.",
			},
			"shape": schema.StringAttribute{
				Computed:    true,
				Description: "The shape of the Exadata infrastructure for the Autonomous VM cluster.",
			},
			"status": schema.StringAttribute{
				CustomType:  status,
				Computed:    true,
				Description: "The status of the Autonomous VM cluster. Possible values include CREATING, AVAILABLE , UPDATING , DELETING , DELETED , FAILED ",
			},
			"status_reason": schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current status of the Autonomous VM cluster.",
			},
			"time_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The time zone of the Autonomous VM cluster. Changing this will force terraform to create new resource.",
			},
			"total_container_databases": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The total number of Autonomous Container Databases that can be created with the allocated local storage. Changing this will force terraform to create new resource.",
			},
			"time_ords_certificate_expires": schema.StringAttribute{
				Computed: true,
			},
			"time_database_ssl_certificate_expires": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration date and time of the database SSL certificate.",
			},
			"maintenance_window": schema.ObjectAttribute{
				Required:   true,
				CustomType: fwtypes.NewObjectTypeOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel](ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Description: "The maintenance window of the Autonomous VM cluster.",
				AttributeTypes: map[string]attr.Type{
					"days_of_week": types.SetType{
						ElemType: fwtypes.StringEnumType[odbtypes.DayOfWeekName](),
					},
					"hours_of_day": types.SetType{
						ElemType: types.Int32Type,
					},
					"lead_time_in_weeks": types.Int32Type,
					"months": types.SetType{
						ElemType: fwtypes.StringEnumType[odbtypes.MonthName](),
					},
					"preference": fwtypes.StringEnumType[odbtypes.PreferenceType](),
					"weeks_of_month": types.SetType{
						ElemType: types.Int32Type,
					},
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCloudAutonomousVmCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateCloudAutonomousVmClusterInput{
		ClientToken:       aws.String(id.UniqueId()),
		Tags:              getTagsIn(ctx),
		MaintenanceWindow: r.expandMaintenanceWindow(ctx, plan.MaintenanceWindow),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCloudAutonomousVmCluster(ctx, &input)
	if err != nil {

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudAutonomousVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CloudAutonomousVmClusterId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudAutonomousVmCluster, plan.DisplayName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdAVMC, err := waitCloudAutonomousVmClusterCreated(ctx, conn, *out.CloudAutonomousVmClusterId, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudAutonomousVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.CreatedAt = types.StringValue(createdAVMC.CreatedAt.Format(time.RFC3339))

	if createdAVMC.TimeDatabaseSslCertificateExpires != nil {
		plan.TimeDatabaseSslCertificateExpires = types.StringValue(createdAVMC.TimeDatabaseSslCertificateExpires.Format(time.RFC3339))
	} else {
		plan.TimeDatabaseSslCertificateExpires = types.StringValue(NotAvailableValues)
	}

	if createdAVMC.TimeOrdsCertificateExpires != nil {
		plan.TimeOrdsCertificateExpires = types.StringValue(createdAVMC.TimeOrdsCertificateExpires.Format(time.RFC3339))
	} else {
		plan.TimeOrdsCertificateExpires = types.StringValue(NotAvailableValues)
	}

	if createdAVMC.MaintenanceWindow != nil {
		plan.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, createdAVMC.MaintenanceWindow)
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, createdAVMC, &plan,
		flex.WithIgnoredFieldNamesAppend("TimeOrdsCertificateExpires"),
		flex.WithIgnoredFieldNamesAppend("TimeDatabaseSslCertificateExpires"))...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCloudAutonomousVmCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().ODBClient(ctx)

	var state cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindCloudAutonomousVmClusterByID(ctx, conn, state.CloudAutonomousVmClusterId.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameCloudAutonomousVmCluster, state.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	if out.TimeOrdsCertificateExpires != nil {
		state.TimeOrdsCertificateExpires = types.StringValue(out.TimeOrdsCertificateExpires.Format(time.RFC3339))
	} else {
		state.TimeOrdsCertificateExpires = types.StringValue(NotAvailableValues)
	}
	if out.TimeDatabaseSslCertificateExpires != nil {
		state.TimeDatabaseSslCertificateExpires = types.StringValue(out.TimeDatabaseSslCertificateExpires.Format(time.RFC3339))
	} else {
		state.TimeDatabaseSslCertificateExpires = types.StringValue(NotAvailableValues)
	}
	state.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, out.MaintenanceWindow)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state,
		flex.WithIgnoredFieldNamesAppend("TimeOrdsCertificateExpires"),
		flex.WithIgnoredFieldNamesAppend("TimeDatabaseSslCertificateExpires"))...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudAutonomousVmCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan, state cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().ODBClient(ctx)

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedAVMC, err := waitCloudAutonomousVmClusterUpdated(ctx, conn, state.CloudAutonomousVmClusterId.ValueString(), updateTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameCloudAutonomousVmCluster, state.CloudAutonomousVmClusterId.String(), err),
			err.Error(),
		)
		return
	}
	plan.CreatedAt = types.StringValue(updatedAVMC.CreatedAt.Format(time.RFC3339))

	if updatedAVMC.TimeDatabaseSslCertificateExpires != nil {
		plan.TimeDatabaseSslCertificateExpires = types.StringValue(updatedAVMC.TimeDatabaseSslCertificateExpires.Format(time.RFC3339))
	} else {
		plan.TimeDatabaseSslCertificateExpires = types.StringValue(NotAvailableValues)
	}

	if updatedAVMC.TimeOrdsCertificateExpires != nil {
		plan.TimeOrdsCertificateExpires = types.StringValue(updatedAVMC.TimeOrdsCertificateExpires.Format(time.RFC3339))
	} else {
		plan.TimeOrdsCertificateExpires = types.StringValue(NotAvailableValues)
	}
	plan.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, updatedAVMC.MaintenanceWindow)
	resp.Diagnostics.Append(flex.Flatten(ctx, updatedAVMC, &plan,
		flex.WithIgnoredFieldNamesAppend("TimeOrdsCertificateExpires"),
		flex.WithIgnoredFieldNamesAppend("TimeDatabaseSslCertificateExpires"))...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCloudAutonomousVmCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().ODBClient(ctx)

	var state cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: state.CloudAutonomousVmClusterId.ValueStringPointer(),
	}

	_, err := conn.DeleteCloudAutonomousVmCluster(ctx, &input)

	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameCloudAutonomousVmCluster, state.CloudAutonomousVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCloudAutonomousVmClusterDeleted(ctx, conn, state.CloudAutonomousVmClusterId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameCloudAutonomousVmCluster, state.CloudAutonomousVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCloudAutonomousVmCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitCloudAutonomousVmClusterCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudAutonomousVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:  enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh: statusCloudAutonomousVmCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudAutonomousVmCluster); ok {
		return out, err
	}

	return nil, err
}

func waitCloudAutonomousVmClusterUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudAutonomousVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusUpdating),
		Target:  enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh: statusCloudAutonomousVmCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudAutonomousVmCluster); ok {
		return out, err
	}

	return nil, err
}

func waitCloudAutonomousVmClusterDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudAutonomousVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusCloudAutonomousVmCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudAutonomousVmCluster); ok {
		return out, err
	}

	return nil, err
}

func statusCloudAutonomousVmCluster(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindCloudAutonomousVmClusterByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindCloudAutonomousVmClusterByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudAutonomousVmCluster, error) {
	input := odb.GetCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: aws.String(id),
	}
	out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.CloudAutonomousVmCluster == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.CloudAutonomousVmCluster, nil
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweeper.go
// as follows:
//
//	awsv2.Register("aws_odb_cloud_autonomous_vm_cluster", sweepCloudAutonomousVmClusters)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepCloudAutonomousVmClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListCloudAutonomousVmClustersInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListCloudAutonomousVmClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.CloudAutonomousVmClusters {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCloudAutonomousVmCluster, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.CloudAutonomousVmClusterId))),
			)
		}
	}

	return sweepResources, nil
}

func (r *resourceCloudAutonomousVmCluster) expandMaintenanceWindow(ctx context.Context, avmcMaintenanceWindowFwTypesObj fwtypes.ObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel]) *odbtypes.MaintenanceWindow {
	var avmcMaintenanceWindowResource cloudAutonomousVmClusterMaintenanceWindowResourceModel

	avmcMaintenanceWindowFwTypesObj.As(ctx, &avmcMaintenanceWindowResource, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})

	var daysOfWeekNames []odbtypes.DayOfWeekName
	avmcMaintenanceWindowResource.DaysOfWeek.ElementsAs(ctx, &daysOfWeekNames, false)
	daysOfWeek := make([]odbtypes.DayOfWeek, 0, len(daysOfWeekNames))

	for _, dayOfWeek := range daysOfWeekNames {
		daysOfWeek = append(daysOfWeek, odbtypes.DayOfWeek{
			Name: dayOfWeek,
		})
	}

	var hoursOfTheDay []int32
	avmcMaintenanceWindowResource.HoursOfDay.ElementsAs(ctx, &hoursOfTheDay, false)

	var monthNames []odbtypes.MonthName
	avmcMaintenanceWindowResource.Months.ElementsAs(ctx, &monthNames, false)
	months := make([]odbtypes.Month, 0, len(monthNames))
	for _, month := range monthNames {
		months = append(months, odbtypes.Month{
			Name: month,
		})
	}

	var weeksOfMonth []int32
	avmcMaintenanceWindowResource.WeeksOfMonth.ElementsAs(ctx, &weeksOfMonth, false)

	odbTypeMW := odbtypes.MaintenanceWindow{
		DaysOfWeek:      daysOfWeek,
		HoursOfDay:      hoursOfTheDay,
		LeadTimeInWeeks: avmcMaintenanceWindowResource.LeadTimeInWeeks.ValueInt32Pointer(),
		Months:          months,
		Preference:      avmcMaintenanceWindowResource.Preference.ValueEnum(),
		WeeksOfMonth:    weeksOfMonth,
	}
	if len(odbTypeMW.DaysOfWeek) == 0 {
		odbTypeMW.DaysOfWeek = nil
	}
	if len(odbTypeMW.HoursOfDay) == 0 {
		odbTypeMW.HoursOfDay = nil
	}
	if len(odbTypeMW.Months) == 0 {
		odbTypeMW.Months = nil
	}
	if len(odbTypeMW.WeeksOfMonth) == 0 {
		odbTypeMW.WeeksOfMonth = nil
	}
	if *odbTypeMW.LeadTimeInWeeks == 0 {
		odbTypeMW.LeadTimeInWeeks = nil
	}
	return &odbTypeMW
}

func (r *resourceCloudAutonomousVmCluster) flattenMaintenanceWindow(ctx context.Context, avmcMW *odbtypes.MaintenanceWindow) fwtypes.ObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel] {
	//days of week
	daysOfWeek := make([]attr.Value, 0, len(avmcMW.DaysOfWeek))
	for _, dayOfWeek := range avmcMW.DaysOfWeek {
		dayOfWeekStringValue := fwtypes.StringEnumValue(dayOfWeek.Name).StringValue
		daysOfWeek = append(daysOfWeek, dayOfWeekStringValue)
	}
	setValueOfDaysOfWeek, _ := basetypes.NewSetValue(types.StringType, daysOfWeek)
	daysOfWeekRead := fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.DayOfWeekName]]{
		SetValue: setValueOfDaysOfWeek,
	}
	//hours of the day
	hoursOfTheDay := make([]attr.Value, 0, len(avmcMW.HoursOfDay))
	for _, hourOfTheDay := range avmcMW.HoursOfDay {
		daysOfWeekInt32Value := types.Int32Value(hourOfTheDay)
		hoursOfTheDay = append(hoursOfTheDay, daysOfWeekInt32Value)
	}
	setValuesOfHoursOfTheDay, _ := basetypes.NewSetValue(types.Int32Type, hoursOfTheDay)
	hoursOfTheDayRead := fwtypes.SetValueOf[types.Int32]{
		SetValue: setValuesOfHoursOfTheDay,
	}
	//monts
	months := make([]attr.Value, 0, len(avmcMW.Months))
	for _, month := range avmcMW.Months {
		monthStringValue := fwtypes.StringEnumValue(month.Name).StringValue
		months = append(months, monthStringValue)
	}
	setValuesOfMonth, _ := basetypes.NewSetValue(types.StringType, months)
	monthsRead := fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.MonthName]]{
		SetValue: setValuesOfMonth,
	}
	//weeks of month
	weeksOfMonth := make([]attr.Value, 0, len(avmcMW.WeeksOfMonth))
	for _, weekOfMonth := range avmcMW.WeeksOfMonth {
		weeksOfMonthInt32Value := types.Int32Value(weekOfMonth)
		weeksOfMonth = append(weeksOfMonth, weeksOfMonthInt32Value)
	}
	setValuesOfWeekOfMonth, _ := basetypes.NewSetValue(types.Int32Type, weeksOfMonth)
	weeksOfMonthRead := fwtypes.SetValueOf[types.Int32]{
		SetValue: setValuesOfWeekOfMonth,
	}

	computedMW := cloudAutonomousVmClusterMaintenanceWindowResourceModel{
		DaysOfWeek:      daysOfWeekRead,
		HoursOfDay:      hoursOfTheDayRead,
		LeadTimeInWeeks: types.Int32PointerValue(avmcMW.LeadTimeInWeeks),
		Months:          monthsRead,
		Preference:      fwtypes.StringEnumValue(avmcMW.Preference),
		WeeksOfMonth:    weeksOfMonthRead,
	}
	if avmcMW.LeadTimeInWeeks == nil {
		computedMW.LeadTimeInWeeks = types.Int32Value(0)
	}
	result, _ := fwtypes.NewObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel](ctx, &computedMW)
	return result
}

type cloudAutonomousVmClusterResourceModel struct {
	framework.WithRegionModel
	CloudAutonomousVmClusterArn                  types.String                                                                  `tfsdk:"arn"`
	CloudAutonomousVmClusterId                   types.String                                                                  `tfsdk:"id"`
	CloudExadataInfrastructureId                 types.String                                                                  `tfsdk:"cloud_exadata_infrastructure_id"`
	AutonomousDataStoragePercentage              types.Float32                                                                 `tfsdk:"autonomous_data_storage_percentage"`
	AutonomousDataStorageSizeInTBs               types.Float64                                                                 `tfsdk:"autonomous_data_storage_size_in_tbs"`
	AvailableAutonomousDataStorageSizeInTBs      types.Float64                                                                 `tfsdk:"available_autonomous_data_storage_size_in_tbs"`
	AvailableContainerDatabases                  types.Int32                                                                   `tfsdk:"available_container_databases"`
	AvailableCpus                                types.Float32                                                                 `tfsdk:"available_cpus"`
	ComputeModel                                 fwtypes.StringEnum[odbtypes.ComputeModel]                                     `tfsdk:"compute_model"`
	CpuCoreCount                                 types.Int32                                                                   `tfsdk:"cpu_core_count"`
	CpuCoreCountPerNode                          types.Int32                                                                   `tfsdk:"cpu_core_count_per_node"`
	CpuPercentage                                types.Float32                                                                 `tfsdk:"cpu_percentage"`
	CreatedAt                                    types.String                                                                  `tfsdk:"created_at" autoflex:",noflatten"`
	DataStorageSizeInGBs                         types.Float64                                                                 `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                         types.Float64                                                                 `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs                       types.Int32                                                                   `tfsdk:"odb_node_storage_size_in_gbs"`
	DbServers                                    fwtypes.SetValueOf[types.String]                                              `tfsdk:"db_servers"`
	Description                                  types.String                                                                  `tfsdk:"description"`
	DisplayName                                  types.String                                                                  `tfsdk:"display_name"`
	Domain                                       types.String                                                                  `tfsdk:"domain"`
	ExadataStorageInTBsLowestScaledValue         types.Float64                                                                 `tfsdk:"exadata_storage_in_tbs_lowest_scaled_value"`
	Hostname                                     types.String                                                                  `tfsdk:"hostname"`
	IsMtlsEnabledVmCluster                       types.Bool                                                                    `tfsdk:"is_mtls_enabled_vm_cluster"`
	LicenseModel                                 fwtypes.StringEnum[odbtypes.LicenseModel]                                     `tfsdk:"license_model"`
	MaxAcdsLowestScaledValue                     types.Int32                                                                   `tfsdk:"max_acds_lowest_scaled_value"`
	MemoryPerOracleComputeUnitInGBs              types.Int32                                                                   `tfsdk:"memory_per_oracle_compute_unit_in_gbs"`
	MemorySizeInGBs                              types.Int32                                                                   `tfsdk:"memory_size_in_gbs"`
	NodeCount                                    types.Int32                                                                   `tfsdk:"node_count"`
	NonProvisionableAutonomousContainerDatabases types.Int32                                                                   `tfsdk:"non_provisionable_autonomous_container_databases"`
	OciResourceAnchorName                        types.String                                                                  `tfsdk:"oci_resource_anchor_name"`
	OciUrl                                       types.String                                                                  `tfsdk:"oci_url"`
	Ocid                                         types.String                                                                  `tfsdk:"ocid"`
	OdbNetworkId                                 types.String                                                                  `tfsdk:"odb_network_id"`
	PercentProgress                              types.Float32                                                                 `tfsdk:"percent_progress"`
	ProvisionableAutonomousContainerDatabases    types.Int32                                                                   `tfsdk:"provisionable_autonomous_container_databases"`
	ProvisionedAutonomousContainerDatabases      types.Int32                                                                   `tfsdk:"provisioned_autonomous_container_databases"`
	ProvisionedCpus                              types.Float32                                                                 `tfsdk:"provisioned_cpus"`
	ReclaimableCpus                              types.Float32                                                                 `tfsdk:"reclaimable_cpus"`
	ReservedCpus                                 types.Float32                                                                 `tfsdk:"reserved_cpus"`
	ScanListenerPortNonTls                       types.Int32                                                                   `tfsdk:"scan_listener_port_non_tls"`
	ScanListenerPortTls                          types.Int32                                                                   `tfsdk:"scan_listener_port_tls"`
	Shape                                        types.String                                                                  `tfsdk:"shape"`
	Status                                       fwtypes.StringEnum[odbtypes.ResourceStatus]                                   `tfsdk:"status"`
	StatusReason                                 types.String                                                                  `tfsdk:"status_reason"`
	TimeZone                                     types.String                                                                  `tfsdk:"time_zone"`
	TotalContainerDatabases                      types.Int32                                                                   `tfsdk:"total_container_databases"`
	Timeouts                                     timeouts.Value                                                                `tfsdk:"timeouts"`
	Tags                                         tftags.Map                                                                    `tfsdk:"tags"`
	TagsAll                                      tftags.Map                                                                    `tfsdk:"tags_all"`
	TimeOrdsCertificateExpires                   types.String                                                                  `tfsdk:"time_ords_certificate_expires"`
	TimeDatabaseSslCertificateExpires            types.String                                                                  `tfsdk:"time_database_ssl_certificate_expires"`
	MaintenanceWindow                            fwtypes.ObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel] `tfsdk:"maintenance_window" autoflex:"-"`
}

type cloudAutonomousVmClusterMaintenanceWindowResourceModel struct {
	DaysOfWeek      fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.DayOfWeekName]] `tfsdk:"days_of_week"`
	HoursOfDay      fwtypes.SetValueOf[types.Int32]                                `tfsdk:"hours_of_day"`
	LeadTimeInWeeks types.Int32                                                    `tfsdk:"lead_time_in_weeks"`
	Months          fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.MonthName]]     `tfsdk:"months"`
	Preference      fwtypes.StringEnum[odbtypes.PreferenceType]                    `tfsdk:"preference"`
	WeeksOfMonth    fwtypes.SetValueOf[types.Int32]                                `tfsdk:"weeks_of_month"`
}
