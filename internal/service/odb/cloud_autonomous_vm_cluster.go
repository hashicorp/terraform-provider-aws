// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
)

type resourceCloudAutonomousVmCluster struct {
	framework.ResourceWithModel[cloudAutonomousVmClusterResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Exadata infrastructure id. Changing this will force terraform to create new resource.",
			},
			"cloud_exadata_infrastructure_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier of the Exadata infrastructure for this VM cluster. Changing this will create a new resource.",
			},
			"autonomous_data_storage_percentage": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Description: "The available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.",
			},
			"available_container_databases": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of Autonomous CDBs that you can create with the currently available storage.",
			},
			"available_cpus": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of CPU cores available for allocation to Autonomous Databases",
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModel,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The compute model of the Autonomous VM cluster: ECPU or OCPU.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: "The percentage of total CPU cores currently in use in the Autonomous VM cluster.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				CustomType:  timetypes.RFC3339Type{},
				Description: "The date and time when the Autonomous VM cluster was created.",
			},
			"data_storage_size_in_gbs": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Description: "The total data storage allocated to the Autonomous VM cluster, in GB.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Description: "The total data storage allocated to the Autonomous VM cluster, in TB.",
			},
			"odb_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
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
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The description of the Autonomous VM cluster.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required:   true,
				Validators: stringLengthBetween1And255Validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The display name of the Autonomous VM cluster. Changing this will force terraform to create new resource.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The domain name of the Autonomous VM cluster.",
			},
			"exadata_storage_in_tbs_lowest_scaled_value": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The total amount of memory allocated to the Autonomous VM cluster, in gigabytes(GB).",
			},
			"node_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of database server nodes in the Autonomous VM cluster.",
			},
			"non_provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of Autonomous CDBs that can't be provisioned because of resource constraints.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the OCI resource anchor associated with this Autonomous VM cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The URL for accessing the OCI console page for this Autonomous VM cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.",
			},
			"odb_network_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier of the ODB network associated with this Autonomous VM Cluster. Changing this will force terraform to create new resource.",
			},
			"odb_network_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier of the ODB network for the VM cluster. This member is required. Changing this will create a new resource.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: `The progress of the current operation on the Autonomous VM cluster, as a percentage.`,
			},
			"provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.",
			},
			"provisioned_autonomous_container_databases": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.",
			},
			"provisioned_cpus": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of CPUs provisioned in the Autonomous VM cluster.",
			},
			"reclaimable_cpus": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: "The number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.",
			},
			"reserved_cpus": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The shape of the Exadata infrastructure for the Autonomous VM cluster.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: status,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The status of the Autonomous VM cluster. Possible values include CREATING, AVAILABLE , UPDATING , DELETING , DELETED , FAILED ",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				CustomType: timetypes.RFC3339Type{},
			},
			"time_database_ssl_certificate_expires": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				CustomType:  timetypes.RFC3339Type{},
				Description: "The expiration date and time of the database SSL certificate.",
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
			"maintenance_window": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Description: "The maintenance window of the Autonomous VM cluster.",

				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"days_of_week": schema.SetAttribute{
							ElementType: fwtypes.NewObjectTypeOf[dayWeekNameAutonomousVmClusterMaintenanceWindowResourceModel](ctx),
							Optional:    true,
							Description: "The days of the week when maintenance can be performed.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"hours_of_day": schema.SetAttribute{
							ElementType: types.Int64Type,
							Optional:    true,
							Description: "The hours of the day when maintenance can be performed.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"lead_time_in_weeks": schema.Int32Attribute{
							Optional:    true,
							Description: "The lead time in weeks before the maintenance window.",
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"months": schema.SetAttribute{
							ElementType: fwtypes.NewObjectTypeOf[monthNameAutonomousVmClusterMaintenanceWindowResourceModel](ctx),
							Optional:    true,
							Description: "The months when maintenance can be performed.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"preference": schema.StringAttribute{
							Required:    true,
							CustomType:  fwtypes.StringEnumType[odbtypes.PreferenceType](),
							Description: "The preference for the maintenance window scheduling.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"weeks_of_month": schema.SetAttribute{
							ElementType: types.Int64Type,
							Optional:    true,
							Description: "Indicates whether to skip release updates during maintenance.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceCloudAutonomousVmCluster) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Neither is present
	if !data.isNetworkARNAndExadataInfraARNPresent() && !data.isNetworkIdAndExadataInfraIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. Neither is present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	//Both are present
	if data.isNetworkARNAndExadataInfraARNPresent() && data.isNetworkIdAndExadataInfraIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. Both are present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	// both exadata infra id and ARN present
	if data.isExadataInfraARNAndIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. exadata infrastructure ID and ARN present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	// both odb network infra and ARN present
	if data.isNetworkARNAndIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. ODB network ID and ARN present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
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
		Tags: getTagsIn(ctx),
	}
	// Handle fallback logic before AutoFlex
	odbNetwork := plan.OdbNetworkId
	if odbNetwork.IsNull() || odbNetwork.IsUnknown() {
		odbNetwork = plan.OdbNetworkArn
	}
	plan.OdbNetworkId = odbNetwork
	cloudExadataInfra := plan.CloudExadataInfrastructureId
	if cloudExadataInfra.IsNull() || cloudExadataInfra.IsUnknown() {
		cloudExadataInfra = plan.CloudExadataInfrastructureArn
	}
	plan.CloudExadataInfrastructureId = cloudExadataInfra
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
	createdAVMC, err := waitCloudAutonomousVmClusterCreated(ctx, conn, aws.ToString(out.CloudAutonomousVmClusterId), createTimeout)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(out.CloudAutonomousVmClusterId))...)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudAutonomousVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, createdAVMC, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCloudAutonomousVmCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state cloudAutonomousVmClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCloudAutonomousVmClusterByID(ctx, conn, state.CloudAutonomousVmClusterId.ValueString())

	if retry.NotFound(err) {
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

func waitCloudAutonomousVmClusterCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudAutonomousVmCluster, error) {
	stateConf := &sdkretry.StateChangeConf{
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

func waitCloudAutonomousVmClusterDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudAutonomousVmCluster, error) {
	stateConf := &sdkretry.StateChangeConf{
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

func statusCloudAutonomousVmCluster(ctx context.Context, conn *odb.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findCloudAutonomousVmClusterByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findCloudAutonomousVmClusterByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudAutonomousVmCluster, error) {
	input := odb.GetCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: aws.String(id),
	}
	out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.CloudAutonomousVmCluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.CloudAutonomousVmCluster, nil
}

type cloudAutonomousVmClusterResourceModel struct {
	framework.WithRegionModel
	CloudAutonomousVmClusterArn                  types.String                                                                            `tfsdk:"arn"`
	CloudAutonomousVmClusterId                   types.String                                                                            `tfsdk:"id"`
	CloudExadataInfrastructureId                 types.String                                                                            `tfsdk:"cloud_exadata_infrastructure_id"`
	CloudExadataInfrastructureArn                types.String                                                                            `tfsdk:"cloud_exadata_infrastructure_arn"`
	AutonomousDataStoragePercentage              types.Float32                                                                           `tfsdk:"autonomous_data_storage_percentage"`
	AutonomousDataStorageSizeInTBs               types.Float64                                                                           `tfsdk:"autonomous_data_storage_size_in_tbs"`
	AvailableAutonomousDataStorageSizeInTBs      types.Float64                                                                           `tfsdk:"available_autonomous_data_storage_size_in_tbs"`
	AvailableContainerDatabases                  types.Int32                                                                             `tfsdk:"available_container_databases"`
	AvailableCpus                                types.Float32                                                                           `tfsdk:"available_cpus"`
	ComputeModel                                 fwtypes.StringEnum[odbtypes.ComputeModel]                                               `tfsdk:"compute_model"`
	CpuCoreCount                                 types.Int32                                                                             `tfsdk:"cpu_core_count"`
	CpuCoreCountPerNode                          types.Int32                                                                             `tfsdk:"cpu_core_count_per_node"`
	CpuPercentage                                types.Float32                                                                           `tfsdk:"cpu_percentage"`
	CreatedAt                                    timetypes.RFC3339                                                                       `tfsdk:"created_at" `
	DataStorageSizeInGBs                         types.Float64                                                                           `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                         types.Float64                                                                           `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs                       types.Int32                                                                             `tfsdk:"odb_node_storage_size_in_gbs"`
	DbServers                                    fwtypes.SetValueOf[types.String]                                                        `tfsdk:"db_servers"`
	Description                                  types.String                                                                            `tfsdk:"description"`
	DisplayName                                  types.String                                                                            `tfsdk:"display_name"`
	Domain                                       types.String                                                                            `tfsdk:"domain"`
	ExadataStorageInTBsLowestScaledValue         types.Float64                                                                           `tfsdk:"exadata_storage_in_tbs_lowest_scaled_value"`
	Hostname                                     types.String                                                                            `tfsdk:"hostname"`
	IsMtlsEnabledVmCluster                       types.Bool                                                                              `tfsdk:"is_mtls_enabled_vm_cluster"`
	LicenseModel                                 fwtypes.StringEnum[odbtypes.LicenseModel]                                               `tfsdk:"license_model"`
	MaxAcdsLowestScaledValue                     types.Int32                                                                             `tfsdk:"max_acds_lowest_scaled_value"`
	MemoryPerOracleComputeUnitInGBs              types.Int32                                                                             `tfsdk:"memory_per_oracle_compute_unit_in_gbs"`
	MemorySizeInGBs                              types.Int32                                                                             `tfsdk:"memory_size_in_gbs"`
	NodeCount                                    types.Int32                                                                             `tfsdk:"node_count"`
	NonProvisionableAutonomousContainerDatabases types.Int32                                                                             `tfsdk:"non_provisionable_autonomous_container_databases"`
	OciResourceAnchorName                        types.String                                                                            `tfsdk:"oci_resource_anchor_name"`
	OciUrl                                       types.String                                                                            `tfsdk:"oci_url"`
	Ocid                                         types.String                                                                            `tfsdk:"ocid"`
	OdbNetworkId                                 types.String                                                                            `tfsdk:"odb_network_id"`
	OdbNetworkArn                                types.String                                                                            `tfsdk:"odb_network_arn"`
	PercentProgress                              types.Float32                                                                           `tfsdk:"percent_progress"`
	ProvisionableAutonomousContainerDatabases    types.Int32                                                                             `tfsdk:"provisionable_autonomous_container_databases"`
	ProvisionedAutonomousContainerDatabases      types.Int32                                                                             `tfsdk:"provisioned_autonomous_container_databases"`
	ProvisionedCpus                              types.Float32                                                                           `tfsdk:"provisioned_cpus"`
	ReclaimableCpus                              types.Float32                                                                           `tfsdk:"reclaimable_cpus"`
	ReservedCpus                                 types.Float32                                                                           `tfsdk:"reserved_cpus"`
	ScanListenerPortNonTls                       types.Int32                                                                             `tfsdk:"scan_listener_port_non_tls"`
	ScanListenerPortTls                          types.Int32                                                                             `tfsdk:"scan_listener_port_tls"`
	Shape                                        types.String                                                                            `tfsdk:"shape"`
	Status                                       fwtypes.StringEnum[odbtypes.ResourceStatus]                                             `tfsdk:"status"`
	StatusReason                                 types.String                                                                            `tfsdk:"status_reason"`
	TimeZone                                     types.String                                                                            `tfsdk:"time_zone"`
	TotalContainerDatabases                      types.Int32                                                                             `tfsdk:"total_container_databases"`
	Timeouts                                     timeouts.Value                                                                          `tfsdk:"timeouts"`
	Tags                                         tftags.Map                                                                              `tfsdk:"tags"`
	TagsAll                                      tftags.Map                                                                              `tfsdk:"tags_all"`
	TimeOrdsCertificateExpires                   timetypes.RFC3339                                                                       `tfsdk:"time_ords_certificate_expires"`
	TimeDatabaseSslCertificateExpires            timetypes.RFC3339                                                                       `tfsdk:"time_database_ssl_certificate_expires"`
	MaintenanceWindow                            fwtypes.ListNestedObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowResourceModel] `tfsdk:"maintenance_window" `
}

type cloudAutonomousVmClusterMaintenanceWindowResourceModel struct {
	DaysOfWeek      fwtypes.SetNestedObjectValueOf[dayWeekNameAutonomousVmClusterMaintenanceWindowResourceModel] `tfsdk:"days_of_week"`
	HoursOfDay      fwtypes.SetValueOf[types.Int64]                                                              `tfsdk:"hours_of_day"`
	LeadTimeInWeeks types.Int32                                                                                  `tfsdk:"lead_time_in_weeks"`
	Months          fwtypes.SetNestedObjectValueOf[monthNameAutonomousVmClusterMaintenanceWindowResourceModel]   `tfsdk:"months"`
	Preference      fwtypes.StringEnum[odbtypes.PreferenceType]                                                  `tfsdk:"preference"`
	WeeksOfMonth    fwtypes.SetValueOf[types.Int64]                                                              `tfsdk:"weeks_of_month"`
}

type dayWeekNameAutonomousVmClusterMaintenanceWindowResourceModel struct {
	Name fwtypes.StringEnum[odbtypes.DayOfWeekName] `tfsdk:"name"`
}

type monthNameAutonomousVmClusterMaintenanceWindowResourceModel struct {
	Name fwtypes.StringEnum[odbtypes.MonthName] `tfsdk:"name"`
}

func (r cloudAutonomousVmClusterResourceModel) isNetworkIdAndExadataInfraIdPresent() bool {
	return !r.OdbNetworkId.IsNull() && !r.CloudExadataInfrastructureId.IsNull()
}

func (r cloudAutonomousVmClusterResourceModel) isNetworkARNAndExadataInfraARNPresent() bool {
	return !r.OdbNetworkArn.IsNull() && !r.CloudExadataInfrastructureArn.IsNull()
}

func (r cloudAutonomousVmClusterResourceModel) isNetworkARNAndIdPresent() bool {
	return !r.OdbNetworkId.IsNull() && !r.OdbNetworkArn.IsNull()
}

func (r cloudAutonomousVmClusterResourceModel) isExadataInfraARNAndIdPresent() bool {
	return !r.CloudExadataInfrastructureId.IsNull() && !r.CloudExadataInfrastructureArn.IsNull()
}
