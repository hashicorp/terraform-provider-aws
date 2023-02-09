package memorydb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCluster() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"engine_patch_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"final_snapshot_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateResourceName(snapshotNameMaxLength),
			},
			"kms_key_arn": {
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateResourceName(clusterNameMaxLength),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateResourceNamePrefix(clusterNameMaxLength - resource.UniqueIDSuffixLength),
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
			"parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"security_group_ids": {
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
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nodes": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      nodeHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"create_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"endpoint": endpointSchema(),
									"name": {
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
			"sns_topic_arn": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
				"address": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"port": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateClusterInput{
		ACLName:                 aws.String(d.Get("acl_name").(string)),
		AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		ClusterName:             aws.String(name),
		NodeType:                aws.String(d.Get("node_type").(string)),
		NumReplicasPerShard:     aws.Int64(int64(d.Get("num_replicas_per_shard").(int))),
		NumShards:               aws.Int64(int64(d.Get("num_shards").(int))),
		Tags:                    Tags(tags.IgnoreAWS()),
		TLSEnabled:              aws.Bool(d.Get("tls_enabled").(bool)),
	}

	if v, ok := d.GetOk("data_tiering"); ok {
		input.DataTiering = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.MaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter_group_name"); ok {
		input.ParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		input.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("snapshot_arns"); ok && len(v.([]interface{})) > 0 {
		v := v.([]interface{})
		input.SnapshotArns = flex.ExpandStringList(v)
		log.Printf("[DEBUG] Restoring MemoryDB Cluster (%s) from S3 snapshots %#v", name, v)
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		input.SnapshotName = aws.String(v.(string))
		log.Printf("[DEBUG] Restoring MemoryDB Cluster (%s) from snapshot %s", name, v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		input.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		input.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		input.SubnetGroupName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating MemoryDB Cluster: %s", input)
	_, err := conn.CreateClusterWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MemoryDB Cluster (%s): %s", name, err)
	}

	if err := waitClusterAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	if d.HasChangesExcept("final_snapshot_name", "tags", "tags_all") {
		waitParameterGroupInSync := false
		waitSecurityGroupsActive := false

		input := &memorydb.UpdateClusterInput{
			ClusterName: aws.String(d.Id()),
		}

		if d.HasChange("acl_name") {
			input.ACLName = aws.String(d.Get("acl_name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("engine_version") {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
		}

		if d.HasChange("maintenance_window") {
			input.MaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		}

		if d.HasChange("node_type") {
			input.NodeType = aws.String(d.Get("node_type").(string))
		}

		if d.HasChange("num_replicas_per_shard") {
			input.ReplicaConfiguration = &memorydb.ReplicaConfigurationRequest{
				ReplicaCount: aws.Int64(int64(d.Get("num_replicas_per_shard").(int))),
			}
		}

		if d.HasChange("num_shards") {
			input.ShardConfiguration = &memorydb.ShardConfigurationRequest{
				ShardCount: aws.Int64(int64(d.Get("num_shards").(int))),
			}
		}

		if d.HasChange("parameter_group_name") {
			input.ParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
			waitParameterGroupInSync = true
		}

		if d.HasChange("security_group_ids") {
			// UpdateCluster reads null and empty slice as "no change", so once
			// at least one security group is present, it's no longer possible
			// to remove all of them.

			v := d.Get("security_group_ids").(*schema.Set)

			if v.Len() == 0 {
				return diag.Errorf("unable to update MemoryDB Cluster (%s): removing all security groups is not possible", d.Id())
			}

			input.SecurityGroupIds = flex.ExpandStringSet(v)
			waitSecurityGroupsActive = true
		}

		if d.HasChange("snapshot_retention_limit") {
			input.SnapshotRetentionLimit = aws.Int64(int64(d.Get("snapshot_retention_limit").(int)))
		}

		if d.HasChange("snapshot_window") {
			input.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		}

		if d.HasChange("sns_topic_arn") {
			v := d.Get("sns_topic_arn").(string)

			input.SnsTopicArn = aws.String(v)

			if v == "" {
				input.SnsTopicStatus = aws.String(ClusterSNSTopicStatusInactive)
			} else {
				input.SnsTopicStatus = aws.String(ClusterSNSTopicStatusActive)
			}
		}

		log.Printf("[DEBUG] Updating MemoryDB Cluster (%s)", d.Id())

		_, err := conn.UpdateClusterWithContext(ctx, input)
		if err != nil {
			return diag.Errorf("error updating MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		if err := waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be modified: %s", d.Id(), err)
		}

		if waitParameterGroupInSync {
			if err := waitClusterParameterGroupInSync(ctx, conn, d.Id()); err != nil {
				return diag.Errorf("error waiting for MemoryDB Cluster (%s) parameter group to be in sync: %s", d.Id(), err)
			}
		}

		if waitSecurityGroupsActive {
			if err := waitClusterSecurityGroupsActive(ctx, conn, d.Id()); err != nil {
				return diag.Errorf("error waiting for MemoryDB Cluster (%s) security groups to be available: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating MemoryDB Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("acl_name", cluster.ACLName)
	d.Set("arn", cluster.ARN)
	d.Set("auto_minor_version_upgrade", cluster.AutoMinorVersionUpgrade)

	if v := cluster.ClusterEndpoint; v != nil {
		d.Set("cluster_endpoint", flattenEndpoint(v))
		d.Set("port", v.Port)
	}

	if v := aws.StringValue(cluster.DataTiering); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return diag.Errorf("error reading data_tiering for MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		d.Set("data_tiering", b)
	}

	d.Set("description", cluster.Description)
	d.Set("engine_patch_version", cluster.EnginePatchVersion)
	d.Set("engine_version", cluster.EngineVersion)
	d.Set("kms_key_arn", cluster.KmsKeyId) // KmsKeyId is actually an ARN here.
	d.Set("maintenance_window", cluster.MaintenanceWindow)
	d.Set("name", cluster.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(cluster.Name)))
	d.Set("node_type", cluster.NodeType)

	numReplicasPerShard, err := deriveClusterNumReplicasPerShard(cluster)
	if err != nil {
		return diag.Errorf("error reading num_replicas_per_shard for MemoryDB Cluster (%s): %s", d.Id(), err)
	}
	d.Set("num_replicas_per_shard", numReplicasPerShard)

	d.Set("num_shards", cluster.NumberOfShards)
	d.Set("parameter_group_name", cluster.ParameterGroupName)

	var securityGroupIds []*string
	for _, v := range cluster.SecurityGroups {
		securityGroupIds = append(securityGroupIds, v.SecurityGroupId)
	}
	d.Set("security_group_ids", flex.FlattenStringSet(securityGroupIds))

	if err := d.Set("shards", flattenShards(cluster.Shards)); err != nil {
		return diag.Errorf("failed to set shards for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)

	if aws.StringValue(cluster.SnsTopicStatus) == ClusterSNSTopicStatusActive {
		d.Set("sns_topic_arn", cluster.SnsTopicArn)
	} else {
		d.Set("sns_topic_arn", "")
	}

	d.Set("subnet_group_name", cluster.SubnetGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	input := &memorydb.DeleteClusterInput{
		ClusterName: aws.String(d.Id()),
	}

	if v := d.Get("final_snapshot_name"); v != nil && len(v.(string)) > 0 {
		input.FinalSnapshotName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting MemoryDB Cluster: (%s)", d.Id())
	_, err := conn.DeleteClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func shardHash(v interface{}) int {
	return create.StringHashcode(v.(map[string]interface{})["name"].(string))
}

func nodeHash(v interface{}) int {
	return create.StringHashcode(v.(map[string]interface{})["name"].(string))
}

func flattenEndpoint(endpoint *memorydb.Endpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := aws.StringValue(endpoint.Address); v != "" {
		m["address"] = v
	}

	if v := aws.Int64Value(endpoint.Port); v != 0 {
		m["port"] = v
	}

	return []interface{}{m}
}

func flattenShards(shards []*memorydb.Shard) *schema.Set {
	shardSet := schema.NewSet(shardHash, nil)

	for _, shard := range shards {
		if shard == nil {
			continue
		}

		nodeSet := schema.NewSet(nodeHash, nil)

		for _, node := range shard.Nodes {
			if node == nil {
				continue
			}

			nodeSet.Add(map[string]interface{}{
				"availability_zone": aws.StringValue(node.AvailabilityZone),
				"create_time":       aws.TimeValue(node.CreateTime).Format(time.RFC3339),
				"endpoint":          flattenEndpoint(node.Endpoint),
				"name":              aws.StringValue(node.Name),
			})
		}

		shardSet.Add(map[string]interface{}{
			"name":      aws.StringValue(shard.Name),
			"num_nodes": int(aws.Int64Value(shard.NumberOfNodes)),
			"nodes":     nodeSet,
			"slots":     aws.StringValue(shard.Slots),
		})
	}

	return shardSet
}

// deriveClusterNumReplicasPerShard determines the replicas per shard
// configuration of a cluster. As this cannot directly be read back, we
// assume that it's the same as that of the largest shard.
//
// For the sake of caution, this search is limited to stable shards.
func deriveClusterNumReplicasPerShard(cluster *memorydb.Cluster) (int, error) {
	var maxNumberOfNodesPerShard int64

	for _, shard := range cluster.Shards {
		if aws.StringValue(shard.Status) != ClusterShardStatusAvailable {
			continue
		}

		n := aws.Int64Value(shard.NumberOfNodes)
		if n > maxNumberOfNodesPerShard {
			maxNumberOfNodesPerShard = n
		}
	}

	if maxNumberOfNodesPerShard == 0 {
		return 0, fmt.Errorf("no available shards found")
	}

	return int(maxNumberOfNodesPerShard - 1), nil
}
