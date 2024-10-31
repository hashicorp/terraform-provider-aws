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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
			Create: schema.DefaultTimeout(clusterAvailableTimeout),
			Update: schema.DefaultTimeout(clusterAvailableTimeout),
			Delete: schema.DefaultTimeout(clusterDeletedTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
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
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"redis",
					"valkey",
				}, false),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Required: true,
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
				Set:      shardHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nodes": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      nodeHash,
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

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateClusterInput{
		ACLName:                 aws.String(d.Get("acl_name").(string)),
		AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
		ClusterName:             aws.String(name),
		Engine:                  aws.String(d.Get(names.AttrEngine).(string)),
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

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.MaintenanceWindow = aws.String(v.(string))
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

	if v, ok := d.GetOk("snapshot_arns"); ok && len(v.([]interface{})) > 0 {
		v := v.([]interface{})
		input.SnapshotArns = flex.ExpandStringValueList(v)
		log.Printf("[DEBUG] Restoring MemoryDB Cluster (%s) from S3 snapshots %#v", name, v)
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		input.SnapshotName = aws.String(v.(string))
		log.Printf("[DEBUG] Restoring MemoryDB Cluster (%s) from snapshot %s", name, v.(string))
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

	log.Printf("[DEBUG] Creating MemoryDB Cluster: %+v", input)
	_, err := conn.CreateCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Cluster (%s): %s", name, err)
	}

	if err := waitClusterAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChangesExcept("final_snapshot_name", names.AttrTags, names.AttrTagsAll) {
		waitParameterGroupInSync := false
		waitSecurityGroupsActive := false

		input := &memorydb.UpdateClusterInput{
			ClusterName: aws.String(d.Id()),
			Engine:      aws.String(d.Get(names.AttrEngine).(string)),
		}

		if d.HasChange("acl_name") {
			input.ACLName = aws.String(d.Get("acl_name").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
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
				input.SnsTopicStatus = aws.String(ClusterSNSTopicStatusInactive)
			} else {
				input.SnsTopicStatus = aws.String(ClusterSNSTopicStatusActive)
			}
		}
		log.Printf("[DEBUG] Updating MemoryDB Cluster (%s)", d.Id())

		_, err := conn.UpdateCluster(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		if err := waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) to be modified: %s", d.Id(), err)
		}

		if waitParameterGroupInSync {
			if err := waitClusterParameterGroupInSync(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) parameter group to be in sync: %s", d.Id(), err)
			}
		}

		if waitSecurityGroupsActive {
			if err := waitClusterSecurityGroupsActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) security groups to be available: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	cluster, err := FindClusterByName(ctx, conn, d.Id())

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
		b, err := strconv.ParseBool(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading data_tiering for MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		d.Set("data_tiering", b)
	}

	d.Set(names.AttrDescription, cluster.Description)
	d.Set("engine_patch_version", cluster.EnginePatchVersion)
	d.Set(names.AttrEngine, cluster.Engine)
	d.Set(names.AttrEngineVersion, cluster.EngineVersion)
	d.Set(names.AttrKMSKeyARN, cluster.KmsKeyId) // KmsKeyId is actually an ARN here.
	d.Set("maintenance_window", cluster.MaintenanceWindow)
	d.Set(names.AttrName, cluster.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(cluster.Name)))
	d.Set("node_type", cluster.NodeType)

	numReplicasPerShard, err := deriveClusterNumReplicasPerShard(cluster)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading num_replicas_per_shard for MemoryDB Cluster (%s): %s", d.Id(), err)
	}
	d.Set("num_replicas_per_shard", numReplicasPerShard)

	d.Set("num_shards", cluster.NumberOfShards)
	d.Set(names.AttrParameterGroupName, cluster.ParameterGroupName)

	var securityGroupIds []*string
	for _, v := range cluster.SecurityGroups {
		securityGroupIds = append(securityGroupIds, v.SecurityGroupId)
	}
	d.Set(names.AttrSecurityGroupIDs, flex.FlattenStringSet(securityGroupIds))

	if err := d.Set("shards", flattenShards(cluster.Shards)); err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to set shards for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)

	if aws.ToString(cluster.SnsTopicStatus) == ClusterSNSTopicStatusActive {
		d.Set(names.AttrSNSTopicARN, cluster.SnsTopicArn)
	} else {
		d.Set(names.AttrSNSTopicARN, "")
	}

	d.Set("subnet_group_name", cluster.SubnetGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)

	return diags
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	input := &memorydb.DeleteClusterInput{
		ClusterName: aws.String(d.Id()),
	}

	if v := d.Get("final_snapshot_name"); v != nil && len(v.(string)) > 0 {
		input.FinalSnapshotName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting MemoryDB Cluster: (%s)", d.Id())
	_, err := conn.DeleteCluster(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func shardHash(v interface{}) int {
	return create.StringHashcode(v.(map[string]interface{})[names.AttrName].(string))
}

func nodeHash(v interface{}) int {
	return create.StringHashcode(v.(map[string]interface{})[names.AttrName].(string))
}

func flattenEndpoint(endpoint *awstypes.Endpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := aws.ToString(endpoint.Address); v != "" {
		m[names.AttrAddress] = v
	}

	if endpoint.Port != 0 {
		m[names.AttrPort] = endpoint.Port
	}

	return []interface{}{m}
}

func flattenShards(shards []awstypes.Shard) *schema.Set {
	shardSet := schema.NewSet(shardHash, nil)

	for _, shard := range shards {
		nodeSet := schema.NewSet(nodeHash, nil)

		for _, node := range shard.Nodes {
			nodeSet.Add(map[string]interface{}{
				names.AttrAvailabilityZone: aws.ToString(node.AvailabilityZone),
				names.AttrCreateTime:       aws.ToTime(node.CreateTime).Format(time.RFC3339),
				names.AttrEndpoint:         flattenEndpoint(node.Endpoint),
				names.AttrName:             aws.ToString(node.Name),
			})
		}

		shardSet.Add(map[string]interface{}{
			names.AttrName: aws.ToString(shard.Name),
			"num_nodes":    int(aws.ToInt32(shard.NumberOfNodes)),
			"nodes":        nodeSet,
			"slots":        aws.ToString(shard.Slots),
		})
	}

	return shardSet
}

// deriveClusterNumReplicasPerShard determines the replicas per shard
// configuration of a cluster. As this cannot directly be read back, we
// assume that it's the same as that of the largest shard.
//
// For the sake of caution, this search is limited to stable shards.
func deriveClusterNumReplicasPerShard(cluster *awstypes.Cluster) (int, error) {
	var maxNumberOfNodesPerShard int32

	for _, shard := range cluster.Shards {
		if aws.ToString(shard.Status) != ClusterShardStatusAvailable {
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
