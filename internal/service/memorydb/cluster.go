// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"acl_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAutoMinorVersionUpgrade: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
					ForceNew: true,
				},
				"cluster_endpoint": endpointSchema(),
				"data_tiering": {
					Type:     schema.TypeBool,
					ForceNew: true,
					Optional: true,
					Default:  false,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Managed by Terraform",
				},
				"engine_patch_version": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"multi_region_cluster_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				names.AttrEngine: {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[clusterEngine](),
				},
				names.AttrEngineVersion: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"final_snapshot_name": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateResourceName(snapshotNameMaxLength),
				},
				names.AttrKMSKeyARN: {
					// The API will accept an ID, but return the ARN on every read.
					// For the sake of consistency, force everyone to use ARN-s.
					// To prevent confusion, the attribute is suffixed _arn rather
					// than the _id implied by the API.
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidARN,
				},
				"maintenance_window": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: verify.ValidOnceAWeekWindowFormat,
				},
				names.AttrName: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrNamePrefix},
					ValidateFunc:  validateResourceName(clusterNameMaxLength),
				},
				names.AttrNamePrefix: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrName},
					ValidateFunc:  validateResourceNamePrefix(clusterNameMaxLength - id.UniqueIDSuffixLength),
				},
				"node_type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"num_replicas_per_shard": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      1,
					ValidateFunc: validation.IntBetween(0, 5),
				},
				"num_shards": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      1,
					ValidateFunc: validation.IntAtLeast(1),
				},
				names.AttrParameterGroupName: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrPort: {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				names.AttrSecurityGroupIDs: {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"shards": {
					Type:     schema.TypeSet,
					Computed: true,
					Set:      clusterShardHash,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Computed: true,
							},
							"nodes": {
								Type:     schema.TypeSet,
								Computed: true,
								Set:      clusterShardNodeHash,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrAvailabilityZone: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrCreateTime: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrEndpoint: endpointSchema(),
										names.AttrName: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"num_nodes": {
								Type:     schema.TypeInt,
								Computed: true,
							},
							"slots": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"snapshot_arns": {
					Type:          schema.TypeList,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"snapshot_name"},
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.All(
							verify.ValidARN,
							validation.StringDoesNotContainAny(","),
						),
					},
				},
				"snapshot_name": {
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"snapshot_arns"},
				},
				"snapshot_retention_limit": {
					Type:         schema.TypeInt,
					Computed:     true,
					Optional:     true,
					ValidateFunc: validation.IntBetween(0, 35),
				},
				"snapshot_window": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: verify.ValidOnceADayWindowFormat,
				},
				names.AttrSNSTopicARN: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
				"subnet_group_name": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"tls_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
					ForceNew: true,
				},
			}
		},
	}
}

func endpointSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrAddress: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrPort: {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := memorydb.CreateClusterInput{
		ACLName:                 aws.String(d.Get("acl_name").(string)),
		AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
		ClusterName:             aws.String(name),
		NodeType:                aws.String(d.Get("node_type").(string)),
		NumReplicasPerShard:     aws.Int32(int32(d.Get("num_replicas_per_shard").(int))),
		NumShards:               aws.Int32(int32(d.Get("num_shards").(int))),
		Tags:                    getTagsIn(ctx),
		TLSEnabled:              aws.Bool(d.Get("tls_enabled").(bool)),
	}

	if v, ok := d.GetOk("data_tiering"); ok {
		input.DataTiering = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.MaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_region_cluster_name"); ok {
		input.MultiRegionClusterName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
		input.ParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("snapshot_arns"); ok && len(v.([]any)) > 0 {
		input.SnapshotArns = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		input.SnapshotName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		input.SnapshotRetentionLimit = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		input.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		input.SubnetGroupName = aws.String(v.(string))
	}

	_, err := conn.CreateCluster(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Cluster (%s): %s", name, err)
	}

	d.SetId(name)

	// If a multi-region cluster name is set, ensure the `aws_memorydb_multi_region_cluster`
	// is created and available before proceeding with cluster creation.
	// Otherwise, the cluster creation will fail.
	if v, ok := d.GetOk("multi_region_cluster_name"); ok {
		if _, err := waitMultiRegionClusterAvailable(ctx, conn, v.(string), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Multi-Region Cluster (%s) create: %s", v.(string), err)
		}
	}

	if _, err := waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	cluster, err := findClusterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("acl_name", cluster.ACLName)
	d.Set(names.AttrARN, cluster.ARN)
	d.Set(names.AttrAutoMinorVersionUpgrade, cluster.AutoMinorVersionUpgrade)
	if v := cluster.ClusterEndpoint; v != nil {
		d.Set("cluster_endpoint", flattenEndpoint(v))
		d.Set(names.AttrPort, v.Port)
	}
	if v := string(cluster.DataTiering); v != "" {
		v, err := strconv.ParseBool(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("data_tiering", v)
	}
	d.Set(names.AttrDescription, cluster.Description)
	d.Set("engine_patch_version", cluster.EnginePatchVersion)
	d.Set("multi_region_cluster_name", cluster.MultiRegionClusterName)
	d.Set(names.AttrEngine, cluster.Engine)
	d.Set(names.AttrEngineVersion, cluster.EngineVersion)
	d.Set(names.AttrKMSKeyARN, cluster.KmsKeyId) // KmsKeyId is actually an ARN here.
	d.Set("maintenance_window", cluster.MaintenanceWindow)
	d.Set(names.AttrName, cluster.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(cluster.Name)))
	d.Set("node_type", cluster.NodeType)
	if v, err := deriveClusterNumReplicasPerShard(cluster); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("num_replicas_per_shard", v)
	}
	d.Set("num_shards", cluster.NumberOfShards)
	d.Set(names.AttrParameterGroupName, cluster.ParameterGroupName)
	d.Set(names.AttrSecurityGroupIDs, tfslices.ApplyToAll(cluster.SecurityGroups, func(v awstypes.SecurityGroupMembership) string {
		return aws.ToString(v.SecurityGroupId)
	}))
	if err := d.Set("shards", flattenShards(cluster.Shards)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting shards: %s", err)
	}
	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)
	if aws.ToString(cluster.SnsTopicStatus) == clusterSNSTopicStatusActive {
		d.Set(names.AttrSNSTopicARN, cluster.SnsTopicArn)
	} else {
		d.Set(names.AttrSNSTopicARN, "")
	}
	d.Set("subnet_group_name", cluster.SubnetGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChangesExcept("final_snapshot_name", names.AttrTags, names.AttrTagsAll) {
		waitParameterGroupInSync := false
		waitSecurityGroupsActive := false

		input := memorydb.UpdateClusterInput{
			ClusterName: aws.String(d.Id()),
		}

		if d.HasChange("acl_name") {
			input.ACLName = aws.String(d.Get("acl_name").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrEngine) {
			input.Engine = aws.String(d.Get(names.AttrEngine).(string))
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("maintenance_window") {
			input.MaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		}

		if d.HasChange("node_type") {
			input.NodeType = aws.String(d.Get("node_type").(string))
		}

		if d.HasChange("num_replicas_per_shard") {
			input.ReplicaConfiguration = &awstypes.ReplicaConfigurationRequest{
				ReplicaCount: int32(d.Get("num_replicas_per_shard").(int)),
			}
		}

		if d.HasChange("num_shards") {
			input.ShardConfiguration = &awstypes.ShardConfigurationRequest{
				ShardCount: int32(d.Get("num_shards").(int)),
			}
		}

		if d.HasChange(names.AttrParameterGroupName) {
			input.ParameterGroupName = aws.String(d.Get(names.AttrParameterGroupName).(string))
			waitParameterGroupInSync = true
		}

		if d.HasChange(names.AttrSecurityGroupIDs) {
			// UpdateCluster reads null and empty slice as "no change", so once
			// at least one security group is present, it's no longer possible
			// to remove all of them.

			v := d.Get(names.AttrSecurityGroupIDs).(*schema.Set)

			if v.Len() == 0 {
				return sdkdiag.AppendErrorf(diags, "unable to update MemoryDB Cluster (%s): removing all security groups is not possible", d.Id())
			}

			input.SecurityGroupIds = flex.ExpandStringValueSet(v)
			waitSecurityGroupsActive = true
		}

		if d.HasChange("snapshot_retention_limit") {
			input.SnapshotRetentionLimit = aws.Int32(int32(d.Get("snapshot_retention_limit").(int)))
		}

		if d.HasChange("snapshot_window") {
			input.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		}

		if d.HasChange(names.AttrSNSTopicARN) {
			v := d.Get(names.AttrSNSTopicARN).(string)
			input.SnsTopicArn = aws.String(v)
			if v == "" {
				input.SnsTopicStatus = aws.String(clusterSNSTopicStatusInactive)
			} else {
				input.SnsTopicStatus = aws.String(clusterSNSTopicStatusActive)
			}
		}

		_, err := conn.UpdateCluster(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) update: %s", d.Id(), err)
		}

		if waitParameterGroupInSync {
			if _, err := waitClusterParameterGroupInSync(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) parameter group: %s", d.Id(), err)
			}
		}

		if waitSecurityGroupsActive {
			if _, err := waitClusterSecurityGroupsActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) security groups: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	input := memorydb.DeleteClusterInput{
		ClusterName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v := d.Get("multi_region_cluster_name"); v != nil && len(v.(string)) > 0 {
		input.MultiRegionClusterName = aws.String(v.(string))
	}

	if v := d.Get("final_snapshot_name"); v != nil && len(v.(string)) > 0 {
		input.FinalSnapshotName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting MemoryDB Cluster: (%s)", d.Get(names.AttrName).(string))
	_, err := conn.DeleteCluster(ctx, &input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Cluster (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Get(names.AttrName).(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) delete: %s", d.Get(names.AttrName).(string), err)
	}

	return diags
}

func findClusterByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Cluster, error) {
	input := memorydb.DescribeClustersInput{
		ClusterName:      aws.String(name),
		ShowShardDetails: aws.Bool(true),
	}

	return findCluster(ctx, conn, &input)
}

func findCluster(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeClustersInput) (*awstypes.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusters(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeClustersInput) ([]awstypes.Cluster, error) {
	var output []awstypes.Cluster

	pages := memorydb.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Clusters...)
	}

	return output, nil
}

func statusCluster(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func statusClusterParameterGroup(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ParameterGroupStatus), nil
	}
}

func statusClusterSecurityGroups(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.SecurityGroups {
			// When at least one security group change is being applied (whether
			// that be adding or removing an SG), say that we're still in progress.
			if aws.ToString(v.Status) != clusterSecurityGroupStatusActive {
				return output, clusterSecurityGroupStatusModifying, nil
			}
		}

		return output, clusterSecurityGroupStatusActive, nil
	}
}

func waitClusterAvailable(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusCreating, clusterStatusUpdating, clusterStatusSnapshotting},
		Target:  []string{clusterStatusAvailable},
		Refresh: statusCluster(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{clusterStatusDeleting},
		Target:       []string{},
		Refresh:      statusCluster(ctx, conn, name),
		Timeout:      timeout,
		Delay:        5 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterParameterGroupInSync(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Cluster, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterParameterGroupStatusApplying},
		Target:  []string{clusterParameterGroupStatusInSync},
		Refresh: statusClusterParameterGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterSecurityGroupsActive(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Cluster, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterSecurityGroupStatusModifying},
		Target:  []string{clusterSecurityGroupStatusActive},
		Refresh: statusClusterSecurityGroups(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

var (
	clusterShardHash     = sdkv2.SimpleSchemaSetFunc(names.AttrName)
	clusterShardNodeHash = sdkv2.SimpleSchemaSetFunc(names.AttrName)
)

func flattenEndpoint(apiObject *awstypes.Endpoint) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if v := aws.ToString(apiObject.Address); v != "" {
		tfMap[names.AttrAddress] = v
	}

	if apiObject.Port != 0 {
		tfMap[names.AttrPort] = apiObject.Port
	}

	return []any{tfMap}
}

func flattenShards(apiObjects []awstypes.Shard) *schema.Set {
	tfSet := schema.NewSet(clusterShardHash, nil)

	for _, apiObject := range apiObjects {
		nodeSet := schema.NewSet(clusterShardNodeHash, nil)

		for _, apiObject := range apiObject.Nodes {
			nodeSet.Add(map[string]any{
				names.AttrAvailabilityZone: aws.ToString(apiObject.AvailabilityZone),
				names.AttrCreateTime:       aws.ToTime(apiObject.CreateTime).Format(time.RFC3339),
				names.AttrEndpoint:         flattenEndpoint(apiObject.Endpoint),
				names.AttrName:             aws.ToString(apiObject.Name),
			})
		}

		tfSet.Add(map[string]any{
			names.AttrName: aws.ToString(apiObject.Name),
			"num_nodes":    aws.ToInt32(apiObject.NumberOfNodes),
			"nodes":        nodeSet,
			"slots":        aws.ToString(apiObject.Slots),
		})
	}

	return tfSet
}

// deriveClusterNumReplicasPerShard determines the replicas per shard
// configuration of a cluster. As this cannot directly be read back, we
// assume that it's the same as that of the largest shard.
//
// For the sake of caution, this search is limited to stable shards.
func deriveClusterNumReplicasPerShard(cluster *awstypes.Cluster) (int, error) {
	var maxNumberOfNodesPerShard int32

	for _, shard := range cluster.Shards {
		if aws.ToString(shard.Status) != clusterShardStatusAvailable {
			continue
		}

		n := aws.ToInt32(shard.NumberOfNodes)
		if n > maxNumberOfNodesPerShard {
			maxNumberOfNodesPerShard = n
		}
	}

	if maxNumberOfNodesPerShard == 0 {
		return 0, fmt.Errorf("no available shards found")
	}

	return int(maxNumberOfNodesPerShard - 1), nil
}
