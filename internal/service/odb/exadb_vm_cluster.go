// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_exadb_vm_cluster", name="ExaDB VM Cluster")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// The generated region override identity test is disabled because Grid Infrastructure image IDs are regional.
// @Testing(identityRegionOverrideTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/odb/types;odbtypes;odbtypes.ExadbVmCluster")
// @Testing(preCheckRegion="us-east-1")
// @Testing(preCheck="testAccPreCheckExaDBVMCluster")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="testAccRandomExaDBVMClusterDisplayName(t)")
// @Testing(requireEnvVarValue="TF_AWS_ODB_EXADB_VM_CLUSTER_GRID_IMAGE_ID")
// @Testing(requireEnvVarValue="TF_AWS_ODB_EXADB_VM_CLUSTER_SSH_PUBLIC_KEY")
func newExaDBVMClusterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &exaDBVMClusterResource{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const ResNameExaDBVMCluster = "ExaDB VM Cluster"

var _ resource.ResourceWithConfigure = (*exaDBVMClusterResource)(nil)

type exaDBVMClusterResource struct {
	framework.ResourceWithModel[exaDBVMClusterResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *exaDBVMClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	licenseModelType := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	gridImageType := fwtypes.StringEnumType[odbtypes.GridImageType]()
	shapeAttributeType := fwtypes.StringEnumType[odbtypes.ShapeAttribute]()
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrClusterName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 11),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`), "must start with a letter and contain only letters, numbers, and hyphens"),
				},
				Description: "Name of the Grid Infrastructure cluster. Changing this value creates a new resource.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Date and time when the ExaDB VM Cluster was created.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`), "must start with a letter or underscore and contain only letters, numbers, underscores, and hyphens"),
				},
				Description: "User-friendly name for the ExaDB VM Cluster.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Domain of the ExaDB VM Cluster.",
			},
			"enabled_ecpu_count": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Number of ECPUs enabled for the ExaDB VM Cluster.",
			},
			"exascale_db_storage_vault_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ARN of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.",
			},
			"exascale_db_storage_vault_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 2048),
				},
				Description: "ID of the Exascale DB Storage Vault for the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"gi_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Oracle Grid Infrastructure software version for the ExaDB VM Cluster.",
			},
			"grid_image_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Grid Infrastructure software image ID for the ExaDB VM Cluster.",
			},
			"grid_image_type": schema.StringAttribute{
				CustomType: gridImageType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Type of Grid Infrastructure image used by the ExaDB VM Cluster.",
			},
			"hostname": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 12),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]$`), "must start with a letter, end with a letter or number, and contain only letters, numbers, and hyphens"),
				},
				Description: "Host name for the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"iam_roles": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterIAMRoleModel](ctx),
				ElementType: types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[exaDBVMClusterIAMRoleModel](ctx)},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "IAM service roles associated with the ExaDB VM Cluster.",
			},
			names.AttrID: framework.IDAttribute(),
			"iorm_config_cache": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterIORMConfigModel](ctx),
				ElementType: types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[exaDBVMClusterIORMConfigModel](ctx)},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "IORM configuration cache details for the ExaDB VM Cluster.",
			},
			"last_update_history_entry_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "OCID of the last maintenance update history entry.",
			},
			"license_model": schema.StringAttribute{
				CustomType: licenseModelType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Oracle license model applied to the ExaDB VM Cluster.",
			},
			"listener_port": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "Listener port configured for the ExaDB VM Cluster.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "Amount of memory allocated to the ExaDB VM Cluster, in GB.",
			},
			"node_count": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
				Description: "Number of nodes in the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"ocid": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "OCID of the ExaDB VM Cluster.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Name of the OCI resource anchor for the ExaDB VM Cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "HTTPS URL of the ExaDB VM Cluster in OCI.",
			},
			"odb_network_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ARN of the ODB network associated with the ExaDB VM Cluster.",
			},
			"odb_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 2048),
				},
				Description: "ID of the ODB network for the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
				Description: "Progress of the current operation, expressed as a percentage.",
			},
			"scan_dns_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "FQDN of the SCAN IP addresses associated with the ExaDB VM Cluster.",
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "OCID of the DNS record for the SCAN IP addresses.",
			},
			"scan_ip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "OCIDs of the SCAN IP addresses associated with the ExaDB VM Cluster.",
			},
			"scan_listener_port_tcp": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.Between(1024, 8999),
				},
				Description: "Port for TCP connections to the SCAN listener. Changing this value creates a new resource.",
			},
			"scan_listener_port_tcp_ssl": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.Between(1024, 8999),
				},
				Description: "Port for SSL/TCP connections to the SCAN listener. Changing this value creates a new resource.",
			},
			"shape": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Shape of the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"shape_attribute": schema.StringAttribute{
				CustomType: shapeAttributeType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Shape attribute for the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"snapshot_file_system_storage": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterStorageDetailsModel](ctx),
				ElementType: types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[exaDBVMClusterStorageDetailsModel](ctx)},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "Snapshot file system storage details for the ExaDB VM Cluster.",
			},
			"ssh_public_keys": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 1024),
				},
				Description: "Public keys used for SSH access to the ExaDB VM Cluster.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: statusType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Current status of the ExaDB VM Cluster.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Additional information about the current ExaDB VM Cluster status.",
			},
			"system_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Operating system version of the image for the ExaDB VM Cluster.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"time_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				Description: "Time zone for the ExaDB VM Cluster. Changing this value creates a new resource.",
			},
			"total_ecpu_count": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(2),
				},
				Description: "Total number of ECPUs for the ExaDB VM Cluster.",
			},
			"total_file_system_storage": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterStorageDetailsModel](ctx),
				ElementType: types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[exaDBVMClusterStorageDetailsModel](ctx)},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "Total file system storage details for the ExaDB VM Cluster.",
			},
			"vip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "OCIDs of the virtual IP addresses associated with the ExaDB VM Cluster.",
			},
			"vm_file_system_storage": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterStorageDetailsModel](ctx),
				ElementType: types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[exaDBVMClusterStorageDetailsModel](ctx)},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "VM file system storage details for the ExaDB VM Cluster.",
			},
			"vm_file_system_storage_total_size_in_gbs": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Total amount of VM file system storage for the ExaDB VM Cluster, in GB.",
			},
		},
		Blocks: map[string]schema.Block{
			"data_collection_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[exaDBVMClusterDataCollectionOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Diagnostic collection preferences for the ExaDB VM Cluster.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"is_diagnostics_events_enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether diagnostic event collection is enabled.",
						},
						"is_health_monitoring_enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether health monitoring is enabled.",
						},
						"is_incident_logs_enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether incident log collection is enabled.",
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

func (r *exaDBVMClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan exaDBVMClusterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var input odb.CreateExadbVmClusterInput
	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateExadbVmCluster(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.DisplayName.String())
		return
	}
	if output == nil || output.ExadbVmClusterId == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.DisplayName.String())
		return
	}

	id := aws.ToString(output.ExadbVmClusterId)
	plan.ExaDBVMClusterID = types.StringValue(id)
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrID), id))
	if response.Diagnostics.HasError() {
		return
	}

	created, err := waitExaDBVMClusterCreated(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ExaDBVMClusterID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, created, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *exaDBVMClusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state exaDBVMClusterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findExaDBVMClusterByID(ctx, conn, state.ExaDBVMClusterID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ExaDBVMClusterID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, state))
}

func (r *exaDBVMClusterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, state exaDBVMClusterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if !diff.HasChanges() {
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
		return
	}

	var input odb.UpdateExadbVmClusterInput
	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...))
	if response.Diagnostics.HasError() {
		return
	}
	input.ExadbVmClusterId = state.ExaDBVMClusterID.ValueStringPointer()

	output, err := conn.UpdateExadbVmCluster(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ExaDBVMClusterID.String())
		return
	}
	if output == nil || output.ExadbVmClusterId == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, state.ExaDBVMClusterID.String())
		return
	}

	updated, err := waitExaDBVMClusterUpdated(ctx, conn, state.ExaDBVMClusterID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ExaDBVMClusterID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, updated, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *exaDBVMClusterResource) flatten(ctx context.Context, exaDBVMCluster *odbtypes.ExadbVmCluster, data *exaDBVMClusterResourceModel) diag.Diagnostics {
	diags := flex.Flatten(ctx, exaDBVMCluster, data, flex.WithFieldNamePrefix("ExadbVmCluster"))
	if exaDBVMCluster.VmFileSystemStorage != nil {
		data.VMFileSystemStorageTotalSizeInGBs = types.Int32PointerValue(exaDBVMCluster.VmFileSystemStorage.TotalSizeInGBs)
	}

	return diags
}

func (r *exaDBVMClusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state exaDBVMClusterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteExadbVmClusterInput{
		ExadbVmClusterId: state.ExaDBVMClusterID.ValueStringPointer(),
	}

	_, err := conn.DeleteExadbVmCluster(ctx, &input)
	if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ExaDBVMClusterID.String())
		return
	}

	_, err = waitExaDBVMClusterDeleted(ctx, conn, state.ExaDBVMClusterID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ExaDBVMClusterID.String())
		return
	}
}

func findExaDBVMClusterByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.ExadbVmCluster, error) {
	input := odb.GetExadbVmClusterInput{
		ExadbVmClusterId: aws.String(id),
	}

	output, err := conn.GetExadbVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}

		return nil, smarterr.NewError(err)
	}

	if output == nil || output.ExadbVmCluster == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.ExadbVmCluster, nil
}

func statusExaDBVMCluster(conn *odb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findExaDBVMClusterByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitExaDBVMClusterCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.ExadbVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:                    enum.Slice(odbtypes.ResourceStatusAvailable),
		Refresh:                   statusExaDBVMCluster(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*odbtypes.ExadbVmCluster); ok {
		if err != nil && aws.ToString(output.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitExaDBVMClusterUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.ExadbVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			odbtypes.ResourceStatusUpdating,
			odbtypes.ResourceStatusMaintenanceInProgress,
		),
		Target:                    enum.Slice(odbtypes.ResourceStatusAvailable),
		Refresh:                   statusExaDBVMCluster(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*odbtypes.ExadbVmCluster); ok {
		if err != nil && aws.ToString(output.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitExaDBVMClusterDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.ExadbVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusExaDBVMCluster(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*odbtypes.ExadbVmCluster); ok {
		if err != nil && aws.ToString(output.StatusReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type exaDBVMClusterResourceModel struct {
	framework.WithRegionModel
	ExaDBVMClusterARN                 types.String                                                              `tfsdk:"arn"`
	ClusterName                       types.String                                                              `tfsdk:"cluster_name"`
	CreatedAt                         timetypes.RFC3339                                                         `tfsdk:"created_at"`
	DataCollectionOptions             fwtypes.ListNestedObjectValueOf[exaDBVMClusterDataCollectionOptionsModel] `tfsdk:"data_collection_options"`
	DisplayName                       types.String                                                              `tfsdk:"display_name"`
	Domain                            types.String                                                              `tfsdk:"domain"`
	EnabledECPUCount                  types.Int32                                                               `tfsdk:"enabled_ecpu_count"`
	ExascaleDBStorageVaultARN         types.String                                                              `tfsdk:"exascale_db_storage_vault_arn"`
	ExascaleDBStorageVaultID          types.String                                                              `tfsdk:"exascale_db_storage_vault_id"`
	ExaDBVMClusterID                  types.String                                                              `tfsdk:"id"`
	GIVersion                         types.String                                                              `tfsdk:"gi_version"`
	GridImageID                       types.String                                                              `tfsdk:"grid_image_id"`
	GridImageType                     fwtypes.StringEnum[odbtypes.GridImageType]                                `tfsdk:"grid_image_type"`
	Hostname                          types.String                                                              `tfsdk:"hostname"`
	IAMRoles                          fwtypes.ListNestedObjectValueOf[exaDBVMClusterIAMRoleModel]               `tfsdk:"iam_roles"`
	IORMConfigCache                   fwtypes.ListNestedObjectValueOf[exaDBVMClusterIORMConfigModel]            `tfsdk:"iorm_config_cache"`
	LastUpdateHistoryEntryID          types.String                                                              `tfsdk:"last_update_history_entry_id"`
	LicenseModel                      fwtypes.StringEnum[odbtypes.LicenseModel]                                 `tfsdk:"license_model"`
	ListenerPort                      types.Int32                                                               `tfsdk:"listener_port"`
	MemorySizeInGBs                   types.Int32                                                               `tfsdk:"memory_size_in_gbs"`
	NodeCount                         types.Int32                                                               `tfsdk:"node_count"`
	OCID                              types.String                                                              `tfsdk:"ocid"`
	OCIResourceAnchorName             types.String                                                              `tfsdk:"oci_resource_anchor_name"`
	OCIURL                            types.String                                                              `tfsdk:"oci_url"`
	ODBNetworkARN                     types.String                                                              `tfsdk:"odb_network_arn"`
	ODBNetworkID                      types.String                                                              `tfsdk:"odb_network_id"`
	PercentProgress                   types.Float32                                                             `tfsdk:"percent_progress"`
	ScanDNSName                       types.String                                                              `tfsdk:"scan_dns_name"`
	ScanDNSRecordID                   types.String                                                              `tfsdk:"scan_dns_record_id"`
	ScanIPIDs                         fwtypes.ListValueOf[types.String]                                         `tfsdk:"scan_ip_ids"`
	ScanListenerPortTCP               types.Int32                                                               `tfsdk:"scan_listener_port_tcp"`
	ScanListenerPortTCPSSL            types.Int32                                                               `tfsdk:"scan_listener_port_tcp_ssl"`
	Shape                             types.String                                                              `tfsdk:"shape"`
	ShapeAttribute                    fwtypes.StringEnum[odbtypes.ShapeAttribute]                               `tfsdk:"shape_attribute"`
	SnapshotFileSystemStorage         fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"snapshot_file_system_storage"`
	SSHPublicKeys                     fwtypes.SetValueOf[types.String]                                          `tfsdk:"ssh_public_keys"`
	Status                            fwtypes.StringEnum[odbtypes.ResourceStatus]                               `tfsdk:"status"`
	StatusReason                      types.String                                                              `tfsdk:"status_reason"`
	SystemVersion                     types.String                                                              `tfsdk:"system_version"`
	Tags                              tftags.Map                                                                `tfsdk:"tags"`
	TagsAll                           tftags.Map                                                                `tfsdk:"tags_all"`
	Timeouts                          timeouts.Value                                                            `tfsdk:"timeouts"`
	TimeZone                          types.String                                                              `tfsdk:"time_zone"`
	TotalECPUCount                    types.Int32                                                               `tfsdk:"total_ecpu_count"`
	TotalFileSystemStorage            fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"total_file_system_storage"`
	VIPIDs                            fwtypes.ListValueOf[types.String]                                         `tfsdk:"vip_ids"`
	VMFileSystemStorage               fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"vm_file_system_storage"`
	VMFileSystemStorageTotalSizeInGBs types.Int32                                                               `tfsdk:"vm_file_system_storage_total_size_in_gbs"`
}

type exaDBVMClusterDataCollectionOptionsModel struct {
	IsDiagnosticsEventsEnabled types.Bool `tfsdk:"is_diagnostics_events_enabled"`
	IsHealthMonitoringEnabled  types.Bool `tfsdk:"is_health_monitoring_enabled"`
	IsIncidentLogsEnabled      types.Bool `tfsdk:"is_incident_logs_enabled"`
}

type exaDBVMClusterIAMRoleModel struct {
	AWSIntegration fwtypes.StringEnum[odbtypes.SupportedAwsIntegration] `tfsdk:"aws_integration"`
	IAMRoleARN     types.String                                         `tfsdk:"iam_role_arn"`
	Status         fwtypes.StringEnum[odbtypes.IamRoleStatus]           `tfsdk:"status"`
	StatusReason   types.String                                         `tfsdk:"status_reason"`
}

type exaDBVMClusterIORMConfigModel struct {
	DBPlans          fwtypes.ListNestedObjectValueOf[exaDBVMClusterDBIORMConfigModel] `tfsdk:"db_plans"`
	LifecycleDetails types.String                                                     `tfsdk:"lifecycle_details"`
	LifecycleState   fwtypes.StringEnum[odbtypes.IormLifecycleState]                  `tfsdk:"lifecycle_state"`
	Objective        fwtypes.StringEnum[odbtypes.Objective]                           `tfsdk:"objective"`
}

type exaDBVMClusterDBIORMConfigModel struct {
	DBName          types.String `tfsdk:"db_name"`
	FlashCacheLimit types.String `tfsdk:"flash_cache_limit"`
	Share           types.Int32  `tfsdk:"share"`
}

type exaDBVMClusterStorageDetailsModel struct {
	TotalSizeInGBs types.Int32 `tfsdk:"total_size_in_gbs"`
}
