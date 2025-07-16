//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"errors"
	awstypes "github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
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
)

var ResourceCloudVmCluster = newResourceCloudVmCluster

type resourceCloudVmCluster struct {
	framework.ResourceWithModel[cloudVmClusterResourceModel]
	framework.WithTimeouts
}

func (r *resourceCloudVmCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	licenseModelType := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	diskRedundancyType := fwtypes.StringEnumType[odbtypes.DiskRedundancy]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu_core_count": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"data_collection_options": schema.ObjectAttribute{
				Computed:   true,
				Optional:   true,
				CustomType: fwtypes.NewObjectTypeOf[cloudVMCResourceModelDataCollectionOptions](ctx),
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"db_servers": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"disk_redundancy": schema.StringAttribute{
				CustomType: diskRedundancyType,
				Computed:   true,
			},
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Computed: true,
			},
			"gi_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname_prefix_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname_prefix": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"iorm_config_cache": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudVMCResourceModelExadataIormConfig](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"lifecycle_details": types.StringType,
						"lifecycle_state":   fwtypes.StringEnumType[odbtypes.IormLifecycleState](),
						"objective":         fwtypes.StringEnumType[odbtypes.Objective](),
						"db_plans": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"db_name":           types.StringType,
									"flash_cache_limit": types.StringType,
									"share":             types.Int32Type,
								},
							},
						},
					},
				},
			},
			"is_local_backup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"is_sparse_diskgroup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"last_update_history_entry_id": schema.StringAttribute{
				Computed: true,
			},
			"license_model": schema.StringAttribute{
				CustomType: licenseModelType,
				Optional:   true,
				Computed:   true,
			},
			"listener_port": schema.Int32Attribute{
				Computed: true,
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"node_count": schema.Int32Attribute{
				Computed: true,
			},
			"ocid": schema.StringAttribute{
				Computed: true,
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
			},
			"odb_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"percent_progress": schema.Float32Attribute{
				Computed: true,
			},
			"scan_dns_name": schema.StringAttribute{
				Computed: true,
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed: true,
			},
			"scan_ip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"shape": schema.StringAttribute{
				Computed: true,
			},
			"ssh_public_keys": schema.SetAttribute{
				Required:    true,
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:   true,
				CustomType: statusType,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
			"storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"system_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"timezone": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"vip_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
			},
			"scan_listener_port_tcp": schema.Int32Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCloudVmCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateCloudVmClusterInput{
		Tags:        getTagsIn(ctx),
		GiVersion:   plan.GiVersion.ValueStringPointer(),
		ClientToken: aws.String(id.UniqueId()),
		Hostname:    plan.HostnamePrefix.ValueStringPointer(),
	}
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
	createdVmCluster, err := waitCloudVmClusterCreated(ctx, conn, *out.CloudVmClusterId, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudVmCluster, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	hostnamePrefix := strings.Split(*input.Hostname, "-")[0]
	plan.HostnamePrefix = types.StringValue(hostnamePrefix)
	plan.HostnamePrefixComputed = types.StringValue(*createdVmCluster.Hostname)
	plan.CreatedAt = types.StringValue(createdVmCluster.CreatedAt.Format(time.RFC3339))
	plan.ScanListenerPortTcp = types.Int32PointerValue(createdVmCluster.ListenerPort)

	resp.Diagnostics.Append(flex.Flatten(ctx, createdVmCluster, &plan, flex.WithIgnoredFieldNamesAppend("HostnamePrefix"),
		flex.WithIgnoredFieldNamesAppend("HostnamePrefixComputed"),
		flex.WithIgnoredFieldNamesAppend("CreatedAt"),
		flex.WithIgnoredFieldNamesAppend("ScanListenerPortTcp"))...)
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

	out, err := FindCloudVmClusterForResourceByID(ctx, conn, state.CloudVmClusterId.ValueString())
	if tfresource.NotFound(err) {
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
	hostnamePrefix := strings.Split(*out.Hostname, "-")[0]
	state.HostnamePrefix = types.StringValue(hostnamePrefix)
	state.HostnamePrefixComputed = types.StringValue(*out.Hostname)
	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	state.ScanListenerPortTcp = types.Int32PointerValue(out.ListenerPort)
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNamesAppend("HostnamePrefix"),
		flex.WithIgnoredFieldNamesAppend("HostnamePrefixComputed"), flex.WithIgnoredFieldNamesAppend("CreatedAt"),
		flex.WithIgnoredFieldNamesAppend("ScanListenerPortTcp"))...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudVmCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan, state cloudVmClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ODBClient(ctx)
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedVMC, err := waitCloudVmClusterUpdated(ctx, conn, plan.CloudVmClusterId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameCloudVmCluster, plan.CloudVmClusterId.String(), err),
			err.Error(),
		)
		return
	}
	hostnamePrefix := strings.Split(*updatedVMC.Hostname, "-")[0]
	plan.HostnamePrefix = types.StringValue(hostnamePrefix)
	plan.HostnamePrefixComputed = types.StringValue(*updatedVMC.Hostname)
	plan.CreatedAt = types.StringValue(updatedVMC.CreatedAt.Format(time.RFC3339))
	plan.ScanListenerPortTcp = types.Int32PointerValue(updatedVMC.ListenerPort)
	resp.Diagnostics.Append(flex.Flatten(ctx, updatedVMC, &plan, flex.WithIgnoredFieldNamesAppend("HostnamePrefix"),
		flex.WithIgnoredFieldNamesAppend("HostnamePrefixComputed"), flex.WithIgnoredFieldNamesAppend("CreatedAt"),
		flex.WithIgnoredFieldNamesAppend("ScanListenerPortTcp"))...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameCloudVmCluster, state.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
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

func (r *resourceCloudVmCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitCloudVmClusterCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudVmCluster, error) {
	stateConf := &retry.StateChangeConf{
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

func waitCloudVmClusterUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudVmCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.ResourceStatusUpdating),
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
	stateConf := &retry.StateChangeConf{
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

func statusCloudVmCluster(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindCloudVmClusterForResourceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindCloudVmClusterForResourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudVmCluster, error) {
	input := odb.GetCloudVmClusterInput{
		CloudVmClusterId: aws.String(id),
	}

	out, err := conn.GetCloudVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.CloudVmCluster == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}
	return out.CloudVmCluster, nil
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
//	awsv2.Register("aws_odb_cloud_vm_cluster", sweepCloudVmClusters)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepCloudVmClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListCloudVmClustersInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListCloudVmClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.CloudVmClusters {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCloudVmCluster, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.CloudVmClusterId))),
			)
		}
	}

	return sweepResources, nil
}

type cloudVmClusterResourceModel struct {
	framework.WithRegionModel
	CloudVmClusterArn            types.String                                                            `tfsdk:"arn"`
	CloudExadataInfrastructureId types.String                                                            `tfsdk:"cloud_exadata_infrastructure_id"`
	CloudVmClusterId             types.String                                                            `tfsdk:"id"`
	ClusterName                  types.String                                                            `tfsdk:"cluster_name"`
	CpuCoreCount                 types.Int32                                                             `tfsdk:"cpu_core_count"`
	DataCollectionOptions        fwtypes.ObjectValueOf[cloudVMCResourceModelDataCollectionOptions]       `tfsdk:"data_collection_options"`
	DataStorageSizeInTBs         types.Float64                                                           `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs       types.Int32                                                             `tfsdk:"db_node_storage_size_in_gbs"`
	DbServers                    fwtypes.SetValueOf[types.String]                                        `tfsdk:"db_servers"`
	DiskRedundancy               fwtypes.StringEnum[odbtypes.DiskRedundancy]                             `tfsdk:"disk_redundancy"`
	DisplayName                  types.String                                                            `tfsdk:"display_name"`
	Domain                       types.String                                                            `tfsdk:"domain"`
	GiVersion                    types.String                                                            `tfsdk:"gi_version"`
	HostnamePrefixComputed       types.String                                                            `tfsdk:"hostname_prefix_computed"`
	HostnamePrefix               types.String                                                            `tfsdk:"hostname_prefix"`
	IormConfigCache              fwtypes.ListNestedObjectValueOf[cloudVMCResourceModelExadataIormConfig] `tfsdk:"iorm_config_cache"`
	IsLocalBackupEnabled         types.Bool                                                              `tfsdk:"is_local_backup_enabled"`
	IsSparseDiskGroupEnabled     types.Bool                                                              `tfsdk:"is_sparse_diskgroup_enabled"`
	LastUpdateHistoryEntryId     types.String                                                            `tfsdk:"last_update_history_entry_id"`
	LicenseModel                 fwtypes.StringEnum[odbtypes.LicenseModel]                               `tfsdk:"license_model"`
	ListenerPort                 types.Int32                                                             `tfsdk:"listener_port"`
	MemorySizeInGbs              types.Int32                                                             `tfsdk:"memory_size_in_gbs"`
	NodeCount                    types.Int32                                                             `tfsdk:"node_count"`
	Ocid                         types.String                                                            `tfsdk:"ocid"`
	OciResourceAnchorName        types.String                                                            `tfsdk:"oci_resource_anchor_name"`
	OciUrl                       types.String                                                            `tfsdk:"oci_url"`
	OdbNetworkId                 types.String                                                            `tfsdk:"odb_network_id"`
	PercentProgress              types.Float32                                                           `tfsdk:"percent_progress"`
	ScanDnsName                  types.String                                                            `tfsdk:"scan_dns_name"`
	ScanDnsRecordId              types.String                                                            `tfsdk:"scan_dns_record_id"`
	ScanIpIds                    fwtypes.ListValueOf[types.String]                                       `tfsdk:"scan_ip_ids"`
	Shape                        types.String                                                            `tfsdk:"shape"`
	SshPublicKeys                fwtypes.SetValueOf[types.String]                                        `tfsdk:"ssh_public_keys"`
	Status                       fwtypes.StringEnum[odbtypes.ResourceStatus]                             `tfsdk:"status"`
	StatusReason                 types.String                                                            `tfsdk:"status_reason"`
	StorageSizeInGBs             types.Int32                                                             `tfsdk:"storage_size_in_gbs"`
	SystemVersion                types.String                                                            `tfsdk:"system_version"`
	Timeouts                     timeouts.Value                                                          `tfsdk:"timeouts"`
	Timezone                     types.String                                                            `tfsdk:"timezone"`
	VipIds                       fwtypes.ListValueOf[types.String]                                       `tfsdk:"vip_ids"`
	CreatedAt                    types.String                                                            `tfsdk:"created_at"`
	ComputeModel                 fwtypes.StringEnum[odbtypes.ComputeModel]                               `tfsdk:"compute_model"`
	ScanListenerPortTcp          types.Int32                                                             `tfsdk:"scan_listener_port_tcp"`
	Tags                         tftags.Map                                                              `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                              `tfsdk:"tags_all"`
}

type cloudVMCResourceModelDataCollectionOptions struct {
	IsDiagnosticsEventsEnabled types.Bool `tfsdk:"is_diagnostics_events_enabled"`
	IsHealthMonitoringEnabled  types.Bool `tfsdk:"is_health_monitoring_enabled"`
	IsIncidentLogsEnabled      types.Bool `tfsdk:"is_incident_logs_enabled"`
}

type cloudVMCResourceModelExadataIormConfig struct {
	DbPlans          fwtypes.ListNestedObjectValueOf[cloudVMCResourceModelDbIormConfig] `tfsdk:"db_plans"`
	LifecycleDetails types.String                                                       `tfsdk:"lifecycle_details"`
	LifecycleState   fwtypes.StringEnum[odbtypes.IormLifecycleState]                    `tfsdk:"lifecycle_state"`
	Objective        fwtypes.StringEnum[odbtypes.Objective]                             `tfsdk:"objective"`
}

type cloudVMCResourceModelDbIormConfig struct {
	DbName          types.String `tfsdk:"db_name"`
	FlashCacheLimit types.String `tfsdk:"flash_cache_limit"`
	Share           types.Int32  `tfsdk:"share"`
}
