// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
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

// @FrameworkResource("aws_odb_cloud_vm_cluster", name="Cloud Vm Cluster")
// @Tags(identifierAttribute="arn")
func newResourceCloudVmCluster(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCloudVmCluster{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameCloudVmCluster = "Cloud Vm Cluster"
	MajorGiVersionPattern = `^\d+\.0\.0\.0$`
	GiVersionSystemTag    = "odb:input_gi_version"
)

var ResourceCloudVmCluster = newResourceCloudVmCluster

type resourceCloudVmCluster struct {
	framework.ResourceWithModel[cloudVmClusterResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceCloudVmCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	licenseModelType := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	diskRedundancyType := fwtypes.StringEnumType[odbtypes.DiskRedundancy]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	giVersionValidator := []validator.String{
		stringvalidator.RegexMatches(regexache.MustCompile(MajorGiVersionPattern), "Gi version must be of the format 19.0.0.0"),
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
				Description: "The unique identifier of the Exadata infrastructure for this VM cluster. Changing this will create a new resource.",
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
			names.AttrClusterName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the Grid Infrastructure (GI) cluster. Changing this will create a new resource.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Description: "The number of CPU cores to enable on the VM cluster. Changing this will create a new resource.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
				Description: "The size of the data disk group, in terabytes (TBs), to allocate for the VM cluster. Changing this will create a new resource.",
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The amount of local node storage, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.",
			},
			"db_servers": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Description: "The list of database servers for the VM cluster. Changing this will create a new resource.",
			},
			"disk_redundancy": schema.StringAttribute{
				CustomType: diskRedundancyType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The type of redundancy for the VM cluster: NORMAL (2-way) or HIGH (3-way).",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "A user-friendly name for the VM cluster. This member is required. Changing this will create a new resource.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The domain name associated with the VM cluster.",
			},
			"gi_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				//Note: underlying API only accepts major gi_version.
				Validators:  giVersionValidator,
				Description: "A valid software version of Oracle Grid Infrastructure (GI). To get the list of valid values, use the ListGiVersions operation and specify the shape of the Exadata infrastructure. Example: 19.0.0.0 This member is required. Changing this will create a new resource.",
			},
			//Underlying API returns complete gi version. For example if gi_version 23.0.0.0 then underlying api returns a version starting with 23
			"gi_version_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "A complete software version of Oracle Grid Infrastructure (GI).",
			},
			//Underlying API treats Hostname as hostname prefix. Therefore, explicitly setting it. API also returns new hostname prefix by appending the input hostname
			//prefix. Therefore, we have hostname_prefix and hostname_prefix_computed
			"hostname_prefix_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The host name for the VM cluster. Constraints: - Can't be \"localhost\" or \"hostname\". - Can't contain \"-version\". - The maximum length of the combined hostname and domain is 63 characters. - The hostname must be unique within the subnet. " +
					"This member is required. Changing this will create a new resource.",
			},
			"hostname_prefix": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The host name prefix for the VM cluster. Constraints: - Can't be \"localhost\" or \"hostname\". - Can't contain \"-version\". - The maximum length of the combined hostname and domain is 63 characters. - The hostname must be unique within the subnet. " +
					"This member is required. Changing this will create a new resource.",
			},
			"iorm_config_cache": schema.ListAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudVMCExadataIormConfigResourceModel](ctx),
				Description: "The Exadata IORM (I/O Resource Manager) configuration cache details for the VM cluster.",
			},
			"is_local_backup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Description: "Specifies whether to enable database backups to local Exadata storage for the VM cluster. Changing this will create a new resource.",
			},
			"is_sparse_diskgroup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Description: "Specifies whether to create a sparse disk group for the VM cluster. Changing this will create a new resource.",
			},
			"last_update_history_entry_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The OCID of the most recent maintenance update history entry.",
			},
			"license_model": schema.StringAttribute{
				CustomType: licenseModelType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The Oracle license model to apply to the VM cluster. Default: LICENSE_INCLUDED. Changing this will create a new resource.",
			},
			"listener_port": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The listener port number configured on the VM cluster.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The amount of memory, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.",
			},
			"node_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The total number of nodes in the VM cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The OCID (Oracle Cloud Identifier) of the VM cluster.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the OCI resource anchor associated with the VM cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The HTTPS link to the VM cluster resource in OCI.",
			},
			"odb_network_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier of the ODB network for the VM cluster. This member is required. Changing this will create a new resource.",
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
				Description: "The percentage of progress made on the current operation for the VM cluster.",
			},
			"scan_dns_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The fully qualified domain name (FQDN) for the SCAN IP addresses associated with the VM cluster.",
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The OCID of the DNS record for the SCAN IPs linked to the VM cluster.",
			},
			"scan_ip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "The list of OCIDs for SCAN IP addresses associated with the VM cluster.",
			},
			"shape": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The hardware model name of the Exadata infrastructure running the VM cluster.",
			},
			"ssh_public_keys": schema.SetAttribute{
				Required:    true,
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Description: "The public key portion of one or more key pairs used for SSH access to the VM cluster. This member is required. Changing this will create a new resource.",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				CustomType:  statusType,
				Description: "The current lifecycle status of the VM cluster.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Additional information regarding the current status of the VM cluster.",
			},
			"storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The local node storage allocated to the VM cluster, in gigabytes (GB).",
			},
			"system_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The operating system version of the image chosen for the VM cluster.",
			},
			"scan_listener_port_tcp": schema.Int32Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The port number for TCP connections to the single client access name (SCAN) listener. " +
					"Valid values: 1024–8999 with the following exceptions: 2484 , 6100 , 6200 , 7060, 7070 , 7085 , and 7879Default: 1521. " +
					"Changing this will create a new resource.",
			},
			"timezone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The configured time zone of the VM cluster. Changing this will create a new resource.",
			},
			"vip_ids": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.ListOfStringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				ElementType: types.StringType,
				Description: "The virtual IP (VIP) addresses assigned to the VM cluster. CRS assigns one VIP per node for failover support.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				CustomType:  timetypes.RFC3339Type{},
				Description: "The timestamp when the VM cluster was created.",
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The compute model used when the instance is created or cloned — either ECPU or OCPU. ECPU is a virtualized compute unit; OCPU is a physical processor core with hyper-threading.",
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
			"data_collection_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudVMCDataCollectionOptionsResourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Description: "The set of preferences for the various diagnostic collection options for the VM cluster. Changing this will create a new resource.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"is_diagnostics_events_enabled": schema.BoolAttribute{
							Required: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						"is_health_monitoring_enabled": schema.BoolAttribute{
							Required: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						"is_incident_logs_enabled": schema.BoolAttribute{
							Required: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceCloudVmCluster) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Neither is present
	if !data.isNetworkARNAndExadataInfraARNPresent() && !data.isNetworkIdAndExadataInfraIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. neither is present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	//Both are present
	if data.isNetworkARNAndExadataInfraARNPresent() && data.isNetworkIdAndExadataInfraIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. both are present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	// both exadata infra id and ARN present
	if data.isExadataInfraARNAndIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. exadata infrastructure id and ARN present")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	// both odb network infra and ARN present
	if data.isNetworkARNAndIdPresent() {
		err := errors.New("either odb_network_id & cloud_exadata_infrastructure_id combination or odb_network_arn & cloud_exadata_infrastructure_arn combination must present. odb network id and ARN ")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	vmcTagAsMap := data.Tags.Elements()
	v, ok := vmcTagAsMap[GiVersionSystemTag]
	if ok {
		if v.String() != data.GiVersion.String() {
			err := errors.New(GiVersionSystemTag + " tag value must be equals to GiVersion value")
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, data.DisplayName.String(), err),
				err.Error(),
			)
			return
		}
	}
}

func (r *resourceCloudVmCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)
	var plan cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	odbNetwork := plan.OdbNetworkId
	if odbNetwork.IsNull() || odbNetwork.IsUnknown() {
		odbNetwork = plan.OdbNetworkArn
	}
	cloudExadataInfra := plan.CloudExadataInfrastructureId
	if cloudExadataInfra.IsNull() || cloudExadataInfra.IsUnknown() {
		cloudExadataInfra = plan.CloudExadataInfrastructureArn
	}
	input := odb.CreateCloudVmClusterInput{
		Tags: getTagsIn(ctx),
		//Underlying API treats Hostname as hostname prefix.
		Hostname: plan.HostnamePrefix.ValueStringPointer(),
	}
	input.OdbNetworkId = odbNetwork.ValueStringPointer()
	input.CloudExadataInfrastructureId = cloudExadataInfra.ValueStringPointer()
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateCloudVmCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CloudVmClusterId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudVmCluster, plan.DisplayName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdVmCluster, err := waitCloudVmClusterCreated(ctx, conn, aws.ToString(out.CloudVmClusterId), createTimeout)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(out.CloudVmClusterId))...)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.HostnamePrefix = flex.StringToFramework(ctx, input.Hostname)
	plan.HostnamePrefixComputed = flex.StringToFramework(ctx, createdVmCluster.Hostname)
	//scan listener port not returned by API directly
	plan.ScanListenerPortTcp = flex.Int32ToFramework(ctx, createdVmCluster.ListenerPort)
	plan.GiVersionComputed = flex.StringToFramework(ctx, createdVmCluster.GiVersion)
	giVersionMajor, err := getMajorGiVersion(ctx, conn, createdVmCluster.CloudVmClusterArn, createdVmCluster.GiVersion)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.GiVersion = flex.StringToFramework(ctx, giVersionMajor)
	plan.OdbNetworkId = flex.StringToFramework(ctx, createdVmCluster.OdbNetworkId)
	plan.OdbNetworkArn = flex.StringToFramework(ctx, createdVmCluster.OdbNetworkArn)
	plan.CloudExadataInfrastructureId = flex.StringToFramework(ctx, createdVmCluster.CloudExadataInfrastructureId)
	plan.CloudExadataInfrastructureArn = flex.StringToFramework(ctx, createdVmCluster.CloudExadataInfrastructureArn)
	resp.Diagnostics.Append(flex.Flatten(ctx, createdVmCluster, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCloudVmCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findCloudVmClusterForResourceByID(ctx, conn, state.CloudVmClusterId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameCloudVmCluster, state.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
	hostnamePrefix := computeHostnamePrefix(out.Hostname)
	state.HostnamePrefix = flex.StringToFramework(ctx, hostnamePrefix)
	state.HostnamePrefixComputed = types.StringValue(*out.Hostname)
	//scan listener port not returned by API directly
	state.ScanListenerPortTcp = flex.Int32ToFramework(ctx, out.ListenerPort)
	state.GiVersionComputed = flex.StringToFramework(ctx, out.GiVersion)
	giVersionMajor, err := getMajorGiVersion(ctx, conn, out.CloudVmClusterArn, out.GiVersion)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudVmCluster, state.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
	state.GiVersion = flex.StringToFramework(ctx, giVersionMajor)
	state.OdbNetworkId = flex.StringToFramework(ctx, out.OdbNetworkId)
	state.OdbNetworkArn = flex.StringToFramework(ctx, out.OdbNetworkArn)
	state.CloudExadataInfrastructureId = flex.StringToFramework(ctx, out.CloudExadataInfrastructureId)
	state.CloudExadataInfrastructureArn = flex.StringToFramework(ctx, out.CloudExadataInfrastructureArn)
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudVmCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteCloudVmClusterInput{
		CloudVmClusterId: state.CloudVmClusterId.ValueStringPointer(),
	}
	_, err := conn.DeleteCloudVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameCloudVmCluster, state.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCloudVmClusterDeleted(ctx, conn, state.CloudVmClusterId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameCloudVmCluster, state.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
}

// computes hostname prefix from hostname prefix computed value.
func computeHostnamePrefix(hostnamePrefixComputed *string) *string {
	suffixIndex := strings.LastIndex(*hostnamePrefixComputed, "-")
	if suffixIndex != -1 {
		actualHostnamePrefix := (*hostnamePrefixComputed)[:suffixIndex]
		return &actualHostnamePrefix
	} else {
		return hostnamePrefixComputed
	}
}
func waitCloudVmClusterCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudVmCluster, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:                    enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:                   statusCloudVmCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudVmCluster); ok {
		return out, err
	}

	return nil, err
}

func waitCloudVmClusterDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudVmCluster, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusCloudVmCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudVmCluster); ok {
		return out, err
	}

	return nil, err
}

func statusCloudVmCluster(ctx context.Context, conn *odb.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findCloudVmClusterForResourceByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findCloudVmClusterForResourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudVmCluster, error) {
	input := odb.GetCloudVmClusterInput{
		CloudVmClusterId: aws.String(id),
	}
	out, err := conn.GetCloudVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.CloudVmCluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	return out.CloudVmCluster, nil
}

// Here we will go through tag to find out whether we can find the input gi_version or not. If not found we will get the version from
// computed gi version to ensure backward compatibility.
func getMajorGiVersion(ctx context.Context, conn *odb.Client, arn *string, giVersionComputed *string) (*string, error) {
	tagsRead, err := listTags(ctx, conn, *arn)
	if err != nil {
		return nil, err
	}
	var inputGiVersion *string
	if tagsRead.KeyExists(GiVersionSystemTag) {
		inputGiVersion = tagsRead.KeyValue(GiVersionSystemTag)
		return inputGiVersion, nil
	} else {
		//This regx based approach is for backward compatibility
		giVersionMajor := strings.Split(*giVersionComputed, ".")[0]
		giVersionMajor = giVersionMajor + ".0.0.0"
		regxGiVersionMajor := regexache.MustCompile(MajorGiVersionPattern)
		if !regxGiVersionMajor.MatchString(giVersionMajor) {
			err := errors.New("gi_version major retrieved from gi_version_computed does not match the pattern 19.0.0.0")
			return nil, err
		}
		return &giVersionMajor, nil
	}
}

type cloudVmClusterResourceModel struct {
	framework.WithRegionModel
	CloudVmClusterArn             types.String                                                                `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String                                                                `tfsdk:"cloud_exadata_infrastructure_id" autoflex:"-"`
	CloudExadataInfrastructureArn types.String                                                                `tfsdk:"cloud_exadata_infrastructure_arn" autoflex:"-"`
	CloudVmClusterId              types.String                                                                `tfsdk:"id"`
	ClusterName                   types.String                                                                `tfsdk:"cluster_name"`
	CpuCoreCount                  types.Int32                                                                 `tfsdk:"cpu_core_count"`
	DataCollectionOptions         fwtypes.ListNestedObjectValueOf[cloudVMCDataCollectionOptionsResourceModel] `tfsdk:"data_collection_options"`
	DataStorageSizeInTBs          types.Float64                                                               `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                                 `tfsdk:"db_node_storage_size_in_gbs"`
	DbServers                     fwtypes.SetValueOf[types.String]                                            `tfsdk:"db_servers"`
	DiskRedundancy                fwtypes.StringEnum[odbtypes.DiskRedundancy]                                 `tfsdk:"disk_redundancy"`
	DisplayName                   types.String                                                                `tfsdk:"display_name"`
	Domain                        types.String                                                                `tfsdk:"domain"`
	GiVersion                     types.String                                                                `tfsdk:"gi_version" autoflex:",noflatten"`
	GiVersionComputed             types.String                                                                `tfsdk:"gi_version_computed" autoflex:",noflatten"`
	HostnamePrefixComputed        types.String                                                                `tfsdk:"hostname_prefix_computed" autoflex:",noflatten"`
	HostnamePrefix                types.String                                                                `tfsdk:"hostname_prefix" autoflex:"-"`
	IormConfigCache               fwtypes.ListNestedObjectValueOf[cloudVMCExadataIormConfigResourceModel]     `tfsdk:"iorm_config_cache"`
	IsLocalBackupEnabled          types.Bool                                                                  `tfsdk:"is_local_backup_enabled"`
	IsSparseDiskGroupEnabled      types.Bool                                                                  `tfsdk:"is_sparse_diskgroup_enabled"`
	LastUpdateHistoryEntryId      types.String                                                                `tfsdk:"last_update_history_entry_id"`
	LicenseModel                  fwtypes.StringEnum[odbtypes.LicenseModel]                                   `tfsdk:"license_model"`
	ListenerPort                  types.Int32                                                                 `tfsdk:"listener_port"`
	MemorySizeInGbs               types.Int32                                                                 `tfsdk:"memory_size_in_gbs"`
	NodeCount                     types.Int32                                                                 `tfsdk:"node_count"`
	Ocid                          types.String                                                                `tfsdk:"ocid"`
	OciResourceAnchorName         types.String                                                                `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                                `tfsdk:"oci_url"`
	OdbNetworkId                  types.String                                                                `tfsdk:"odb_network_id" autoflex:"-"`
	OdbNetworkArn                 types.String                                                                `tfsdk:"odb_network_arn" autoflex:"-"`
	PercentProgress               types.Float32                                                               `tfsdk:"percent_progress"`
	ScanDnsName                   types.String                                                                `tfsdk:"scan_dns_name"`
	ScanDnsRecordId               types.String                                                                `tfsdk:"scan_dns_record_id"`
	ScanIpIds                     fwtypes.ListValueOf[types.String]                                           `tfsdk:"scan_ip_ids"`
	Shape                         types.String                                                                `tfsdk:"shape"`
	SshPublicKeys                 fwtypes.SetValueOf[types.String]                                            `tfsdk:"ssh_public_keys"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                                 `tfsdk:"status"`
	StatusReason                  types.String                                                                `tfsdk:"status_reason"`
	StorageSizeInGBs              types.Int32                                                                 `tfsdk:"storage_size_in_gbs"`
	SystemVersion                 types.String                                                                `tfsdk:"system_version"`
	Timeouts                      timeouts.Value                                                              `tfsdk:"timeouts"`
	Timezone                      types.String                                                                `tfsdk:"timezone"`
	VipIds                        fwtypes.ListValueOf[types.String]                                           `tfsdk:"vip_ids"`
	CreatedAt                     timetypes.RFC3339                                                           `tfsdk:"created_at"`
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                                   `tfsdk:"compute_model"`
	ScanListenerPortTcp           types.Int32                                                                 `tfsdk:"scan_listener_port_tcp" autoflex:",noflatten"`
	Tags                          tftags.Map                                                                  `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                                  `tfsdk:"tags_all"`
}

type cloudVMCDataCollectionOptionsResourceModel struct {
	IsDiagnosticsEventsEnabled types.Bool `tfsdk:"is_diagnostics_events_enabled"`
	IsHealthMonitoringEnabled  types.Bool `tfsdk:"is_health_monitoring_enabled"`
	IsIncidentLogsEnabled      types.Bool `tfsdk:"is_incident_logs_enabled"`
}

type cloudVMCExadataIormConfigResourceModel struct {
	DbPlans          fwtypes.ListNestedObjectValueOf[cloudVMCDbIormConfigResourceModel] `tfsdk:"db_plans"`
	LifecycleDetails types.String                                                       `tfsdk:"lifecycle_details"`
	LifecycleState   fwtypes.StringEnum[odbtypes.IormLifecycleState]                    `tfsdk:"lifecycle_state"`
	Objective        fwtypes.StringEnum[odbtypes.Objective]                             `tfsdk:"objective"`
}

type cloudVMCDbIormConfigResourceModel struct {
	DbName          types.String `tfsdk:"db_name"`
	FlashCacheLimit types.String `tfsdk:"flash_cache_limit"`
	Share           types.Int32  `tfsdk:"share"`
}

func (r cloudVmClusterResourceModel) isNetworkIdAndExadataInfraIdPresent() bool {
	return !r.OdbNetworkId.IsNull() && !r.CloudExadataInfrastructureId.IsNull()
}

func (r cloudVmClusterResourceModel) isNetworkARNAndExadataInfraARNPresent() bool {
	return !r.OdbNetworkArn.IsNull() && !r.CloudExadataInfrastructureArn.IsNull()
}

func (r cloudVmClusterResourceModel) isNetworkARNAndIdPresent() bool {
	return !r.OdbNetworkId.IsNull() && !r.OdbNetworkArn.IsNull()
}

func (r cloudVmClusterResourceModel) isExadataInfraARNAndIdPresent() bool {
	return !r.CloudExadataInfrastructureId.IsNull() && !r.CloudExadataInfrastructureArn.IsNull()
}
