// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_replication_group", name="Replication Group")
// @Tags(identifierAttribute="arn")
func ResourceReplicationGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationGroupCreate,
		ReadWithoutTimeout:   resourceReplicationGroupRead,
		UpdateWithoutTimeout: resourceReplicationGroupUpdate,
		DeleteWithoutTimeout: resourceReplicationGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"auth_token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ValidateFunc:  validReplicationGroupAuthToken,
				ConflictsWith: []string{"user_group_ids"},
			},
			"auth_token_update_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(elasticache.AuthTokenUpdateStrategyType_Values(), true),
				Default:      elasticache.AuthTokenUpdateStrategyTypeRotate,
			},
			"auto_minor_version_upgrade": {
				Type:         nullable.TypeNullableBool,
				Optional:     true,
				Computed:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableBool,
			},
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"configuration_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_tiering_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineRedis,
				ValidateFunc: validation.StringInSlice([]string{engineRedis}, true),
			},
			"engine_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validRedisVersionString,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ConflictsWith: []string{
					"num_node_groups",
					"parameter_group_name",
					"engine",
					"engine_version",
					"node_type",
					"security_group_names",
					"transit_encryption_enabled",
					"at_rest_encryption_enabled",
					"snapshot_arns",
					"snapshot_name",
				},
			},
			"ip_discovery": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(elasticache.IpDiscovery_Values(), false),
			},
			"log_delivery_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticache.DestinationType_Values(), false),
						},
						"destination": {
							Type:     schema.TypeString,
							Required: true,
						},
						"log_format": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticache.LogFormat_Values(), false),
						},
						"log_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticache.LogType_Values(), false),
						},
					},
				},
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					// ElastiCache always changes the maintenance to lowercase
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"member_clusters": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"multi_az_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(elasticache.NetworkType_Values(), false),
			},
			"node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"notification_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"num_cache_clusters": {
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"num_node_groups"},
			},
			"num_node_groups": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"num_cache_clusters", "global_replication_group_id"},
			},
			"parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.HasPrefix(old, "global-datastore-")
				},
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress default Redis ports when not defined
					if !d.IsNewResource() && new == "0" && old == defaultRedisPort {
						return true
					}
					return false
				},
			},
			"preferred_cache_cluster_azs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"primary_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reader_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replicas_per_node_group": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"replication_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateReplicationGroupID,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"snapshot_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				// Note: Unlike aws_elasticache_cluster, this does not have a limit of 1 item.
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						verify.ValidARN,
						validation.StringDoesNotContainAny(","),
					),
				},
				Set: schema.HashString,
			},
			"snapshot_retention_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtMost(35),
			},
			"snapshot_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"user_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				ConflictsWith: []string{"auth_token"},
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		SchemaVersion: 2,
		// SchemaVersion: 1 did not include any state changes via MigrateState.
		// Perform a no-operation state upgrade for Terraform 0.12 compatibility.
		// Future state migrations should be performed with StateUpgraders.
		MigrateState: func(v int, inst *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
			return inst, nil
		},

		StateUpgraders: []schema.StateUpgrader{
			// v5.27.0 introduced the auth_token_update_strategy argument with a default
			// value required to preserve backward compatibility. In order to prevent
			// differences and attempted modifications on upgrade, the default value
			// must be written to state via a state upgrader.
			{
				Type:    resourceReplicationGroupConfigV1().CoreConfigSchema().ImpliedType(),
				Upgrade: replicationGroupStateUpgradeV1,
				Version: 1,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ReplicationGroupDefaultCreatedTimeout),
			Delete: schema.DefaultTimeout(ReplicationGroupDefaultDeletedTimeout),
			Update: schema.DefaultTimeout(ReplicationGroupDefaultUpdatedTimeout),
		},

		CustomizeDiff: customdiff.Sequence(
			CustomizeDiffValidateReplicationGroupAutomaticFailover,
			customizeDiffEngineVersionForceNewOnDowngrade,
			customdiff.ComputedIf("member_clusters", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("num_cache_clusters") ||
					diff.HasChange("num_node_groups") ||
					diff.HasChange("replicas_per_node_group")
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceReplicationGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	replicationGroupID := d.Get("replication_group_id").(string)
	input := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ReplicationGroupDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_tiering_enabled"); ok {
		input.DataTieringEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		input.GlobalReplicationGroupId = aws.String(v.(string))
	} else {
		// This cannot be handled at plan-time
		nodeType := d.Get("node_type").(string)
		if nodeType == "" {
			return sdkdiag.AppendErrorf(diags, `"node_type" is required unless "global_replication_group_id" is set.`)
		}
		input.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		input.CacheNodeType = aws.String(nodeType)
		input.Engine = aws.String(d.Get("engine").(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_minor_version_upgrade"); ok {
		if v, null, _ := nullable.Bool(v.(string)).Value(); !null {
			input.AutoMinorVersionUpgrade = aws.Bool(v)
		}
	}

	if preferredAZs, ok := d.GetOk("preferred_cache_cluster_azs"); ok {
		input.PreferredCacheClusterAZs = flex.ExpandStringList(preferredAZs.([]interface{}))
	}

	if v, ok := d.GetOk("parameter_group_name"); ok {
		input.CacheParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_discovery"); ok {
		input.IpDiscovery = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_type"); ok {
		input.NetworkType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		input.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		input.CacheSubnetGroupName = aws.String(v.(string))
	}

	if SGNames := d.Get("security_group_names").(*schema.Set); SGNames.Len() > 0 {
		input.CacheSecurityGroupNames = flex.ExpandStringSet(SGNames)
	}

	if SGIds := d.Get("security_group_ids").(*schema.Set); SGIds.Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(SGIds)
	}

	if snaps := d.Get("snapshot_arns").(*schema.Set); snaps.Len() > 0 {
		input.SnapshotArns = flex.ExpandStringSet(snaps)
	}

	if v, ok := d.GetOk("log_delivery_configuration"); ok {
		input.LogDeliveryConfigurations = []*elasticache.LogDeliveryConfigurationRequest{}
		v := v.(*schema.Set).List()
		for _, v := range v {
			logDeliveryConfigurationRequest := expandLogDeliveryConfigurations(v.(map[string]interface{}))
			input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
		}
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if _, ok := d.GetOk("multi_az_enabled"); ok {
		input.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		input.NotificationTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		input.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		input.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		input.SnapshotName = aws.String(v.(string))
	}

	if _, ok := d.GetOk("transit_encryption_enabled"); ok {
		input.TransitEncryptionEnabled = aws.Bool(d.Get("transit_encryption_enabled").(bool))
	}

	if _, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		input.AtRestEncryptionEnabled = aws.Bool(d.Get("at_rest_encryption_enabled").(bool))
	}

	if v, ok := d.GetOk("auth_token"); ok {
		input.AuthToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("num_node_groups"); ok && v != 0 {
		input.NumNodeGroups = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("replicas_per_node_group"); ok {
		input.ReplicasPerNodeGroup = aws.Int64(int64(v.(int)))
	}

	if numCacheClusters, ok := d.GetOk("num_cache_clusters"); ok {
		input.NumCacheClusters = aws.Int64(int64(numCacheClusters.(int)))
	}

	if userGroupIds := d.Get("user_group_ids").(*schema.Set); userGroupIds.Len() > 0 {
		input.UserGroupIds = flex.ExpandStringSet(userGroupIds)
	}

	output, err := conn.CreateReplicationGroupWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateReplicationGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache Replication Group (%s): %s", replicationGroupID, err)
	}

	d.SetId(aws.StringValue(output.ReplicationGroup.ReplicationGroupId))

	if _, err := WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		// When adding a replication group to a global replication group, the replication group can be in the "available"
		// state, but the global replication group can still be in the "modifying" state. Wait for the replication group
		// to be fully added to the global replication group.
		// API calls to the global replication group can be made in any region.
		if _, err := waitGlobalReplicationGroupAvailable(ctx, conn, v.(string), globalReplicationGroupDefaultCreatedTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Global Replication Group (%s) to be available: %s", v, err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.ReplicationGroup.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceReplicationGroupRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ElastiCache Replication Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicationGroupRead(ctx, d, meta)...)
}

func resourceReplicationGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	rgp, err := FindReplicationGroupByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): %s", d.Id(), err)
	}

	if aws.StringValue(rgp.Status) == ReplicationGroupStatusDeleting {
		log.Printf("[WARN] ElastiCache Replication Group (%s) is currently in the `deleting` status, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if rgp.GlobalReplicationGroupInfo != nil && rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId != nil {
		d.Set("global_replication_group_id", rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
	}

	if rgp.AutomaticFailover != nil {
		switch strings.ToLower(aws.StringValue(rgp.AutomaticFailover)) {
		case elasticache.AutomaticFailoverStatusDisabled, elasticache.AutomaticFailoverStatusDisabling:
			d.Set("automatic_failover_enabled", false)
		case elasticache.AutomaticFailoverStatusEnabled, elasticache.AutomaticFailoverStatusEnabling:
			d.Set("automatic_failover_enabled", true)
		default:
			log.Printf("Unknown AutomaticFailover state %q", aws.StringValue(rgp.AutomaticFailover))
		}
	}

	if rgp.MultiAZ != nil {
		switch strings.ToLower(aws.StringValue(rgp.MultiAZ)) {
		case elasticache.MultiAZStatusEnabled:
			d.Set("multi_az_enabled", true)
		case elasticache.MultiAZStatusDisabled:
			d.Set("multi_az_enabled", false)
		default:
			log.Printf("Unknown MultiAZ state %q", aws.StringValue(rgp.MultiAZ))
		}
	}

	d.Set("kms_key_id", rgp.KmsKeyId)
	d.Set("description", rgp.Description)
	d.Set("num_cache_clusters", len(rgp.MemberClusters))
	if err := d.Set("member_clusters", flex.FlattenStringSet(rgp.MemberClusters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting member_clusters: %s", err)
	}

	d.Set("num_node_groups", len(rgp.NodeGroups))
	d.Set("replicas_per_node_group", len(rgp.NodeGroups[0].NodeGroupMembers)-1)

	d.Set("cluster_enabled", rgp.ClusterEnabled)
	d.Set("replication_group_id", rgp.ReplicationGroupId)
	d.Set("arn", rgp.ARN)
	d.Set("data_tiering_enabled", aws.StringValue(rgp.DataTiering) == elasticache.DataTieringStatusEnabled)

	d.Set("ip_discovery", rgp.IpDiscovery)
	d.Set("network_type", rgp.NetworkType)

	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(rgp.LogDeliveryConfigurations))
	d.Set("snapshot_window", rgp.SnapshotWindow)
	d.Set("snapshot_retention_limit", rgp.SnapshotRetentionLimit)

	if rgp.ConfigurationEndpoint != nil {
		d.Set("port", rgp.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint_address", rgp.ConfigurationEndpoint.Address)
	} else {
		log.Printf("[DEBUG] ElastiCache Replication Group (%s) Configuration Endpoint is nil", d.Id())

		if rgp.NodeGroups[0].PrimaryEndpoint != nil {
			log.Printf("[DEBUG] ElastiCache Replication Group (%s) Primary Endpoint is not nil", d.Id())
			d.Set("port", rgp.NodeGroups[0].PrimaryEndpoint.Port)
			d.Set("primary_endpoint_address", rgp.NodeGroups[0].PrimaryEndpoint.Address)
		}

		if rgp.NodeGroups[0].ReaderEndpoint != nil {
			d.Set("reader_endpoint_address", rgp.NodeGroups[0].ReaderEndpoint.Address)
		}
	}

	d.Set("user_group_ids", rgp.UserGroupIds)

	// Tags cannot be read when the replication group is not Available
	log.Printf("[DEBUG] Waiting for ElastiCache Replication Group (%s) to become available", d.Id())

	_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group to be available (%s): %s", aws.StringValue(rgp.ARN), err)
	}

	log.Printf("[DEBUG] ElastiCache Replication Group (%s): Checking underlying cache clusters", d.Id())

	// This section reads settings that require checking the underlying cache clusters
	if rgp.NodeGroups != nil && rgp.NodeGroups[0] != nil && len(rgp.NodeGroups[0].NodeGroupMembers) != 0 {
		cacheCluster := rgp.NodeGroups[0].NodeGroupMembers[0]

		res, err := conn.DescribeCacheClustersWithContext(ctx, &elasticache.DescribeCacheClustersInput{
			CacheClusterId:    cacheCluster.CacheClusterId,
			ShowCacheNodeInfo: aws.Bool(true),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): reading Cache Cluster (%s): %s", d.Id(), aws.StringValue(cacheCluster.CacheClusterId), err)
		}

		if len(res.CacheClusters) == 0 {
			return diags
		}

		c := res.CacheClusters[0]

		if err := setFromCacheCluster(d, c); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): reading Cache Cluster (%s): %s", d.Id(), aws.StringValue(cacheCluster.CacheClusterId), err)
		}

		d.Set("at_rest_encryption_enabled", c.AtRestEncryptionEnabled)
		d.Set("transit_encryption_enabled", c.TransitEncryptionEnabled)

		if c.AuthTokenEnabled != nil && !aws.BoolValue(c.AuthTokenEnabled) {
			d.Set("auth_token", nil)
		}
	}

	return diags
}

func resourceReplicationGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		if d.HasChanges(
			"num_node_groups",
			"replicas_per_node_group",
		) {
			err := modifyReplicationGroupShardConfiguration(ctx, conn, d)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ElastiCache Replication Group (%s) shard configuration: %s", d.Id(), err)
			}
		} else if d.HasChange("num_cache_clusters") {
			err := modifyReplicationGroupNumCacheClusters(ctx, conn, d, "num_cache_clusters")
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ElastiCache Replication Group (%s) clusters: %s", d.Id(), err)
			}
		}

		requestUpdate := false
		input := &elasticache.ModifyReplicationGroupInput{
			ApplyImmediately:   aws.Bool(d.Get("apply_immediately").(bool)),
			ReplicationGroupId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.ReplicationGroupDescription = aws.String(d.Get("description").(string))
			requestUpdate = true
		}

		if d.HasChange("ip_discovery") {
			input.IpDiscovery = aws.String(d.Get("ip_discovery").(string))
			requestUpdate = true
		}

		if d.HasChange("network_type") {
			input.IpDiscovery = aws.String(d.Get("network_type").(string))
			requestUpdate = true
		}

		if d.HasChange("automatic_failover_enabled") {
			input.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
			requestUpdate = true
		}

		if d.HasChange("auto_minor_version_upgrade") {
			v := d.Get("auto_minor_version_upgrade")
			if v, null, _ := nullable.Bool(v.(string)).Value(); !null {
				input.AutoMinorVersionUpgrade = aws.Bool(v)
			}
			requestUpdate = true
		}

		if d.HasChange("security_group_ids") {
			if attr := d.Get("security_group_ids").(*schema.Set); attr.Len() > 0 {
				input.SecurityGroupIds = flex.ExpandStringSet(attr)
				requestUpdate = true
			}
		}

		if d.HasChange("security_group_names") {
			if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
				input.CacheSecurityGroupNames = flex.ExpandStringSet(attr)
				requestUpdate = true
			}
		}

		if d.HasChange("log_delivery_configuration") {
			oldLogDeliveryConfig, newLogDeliveryConfig := d.GetChange("log_delivery_configuration")

			input.LogDeliveryConfigurations = []*elasticache.LogDeliveryConfigurationRequest{}
			logTypesToSubmit := make(map[string]bool)

			currentLogDeliveryConfig := newLogDeliveryConfig.(*schema.Set).List()
			for _, current := range currentLogDeliveryConfig {
				logDeliveryConfigurationRequest := expandLogDeliveryConfigurations(current.(map[string]interface{}))
				logTypesToSubmit[*logDeliveryConfigurationRequest.LogType] = true
				input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
			}

			previousLogDeliveryConfig := oldLogDeliveryConfig.(*schema.Set).List()
			for _, previous := range previousLogDeliveryConfig {
				logDeliveryConfigurationRequest := expandEmptyLogDeliveryConfigurations(previous.(map[string]interface{}))
				//if something was removed, send an empty request
				if !logTypesToSubmit[*logDeliveryConfigurationRequest.LogType] {
					input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
				}
			}
			requestUpdate = true
		}

		if d.HasChange("maintenance_window") {
			input.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
			requestUpdate = true
		}

		if d.HasChange("multi_az_enabled") {
			input.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
			requestUpdate = true
		}

		if d.HasChange("notification_topic_arn") {
			input.NotificationTopicArn = aws.String(d.Get("notification_topic_arn").(string))
			requestUpdate = true
		}

		if d.HasChange("parameter_group_name") {
			input.CacheParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
			requestUpdate = true
		}

		if d.HasChange("engine_version") {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
			requestUpdate = true
		}

		if d.HasChange("snapshot_retention_limit") {
			// This is a real hack to set the Snapshotting Cluster ID to be the first Cluster in the RG
			o, _ := d.GetChange("snapshot_retention_limit")
			if o.(int) == 0 {
				input.SnapshottingClusterId = aws.String(fmt.Sprintf("%s-001", d.Id()))
			}

			input.SnapshotRetentionLimit = aws.Int64(int64(d.Get("snapshot_retention_limit").(int)))
			requestUpdate = true
		}

		if d.HasChange("snapshot_window") {
			input.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
			requestUpdate = true
		}

		if d.HasChange("node_type") {
			input.CacheNodeType = aws.String(d.Get("node_type").(string))
			requestUpdate = true
		}

		if d.HasChange("user_group_ids") {
			old, new := d.GetChange("user_group_ids")
			newSet := new.(*schema.Set)
			oldSet := old.(*schema.Set)
			add := newSet.Difference(oldSet)
			remove := oldSet.Difference(newSet)

			if add.Len() > 0 {
				input.UserGroupIdsToAdd = flex.ExpandStringSet(add)
				requestUpdate = true
			}

			if remove.Len() > 0 {
				input.UserGroupIdsToRemove = flex.ExpandStringSet(remove)
				requestUpdate = true
			}
		}

		if requestUpdate {
			_, err := conn.ModifyReplicationGroupWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating ElastiCache Replication Group (%s): %s", d.Id(), err)
			}

			_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) to update: %s", d.Id(), err)
			}
		}

		if d.HasChanges("auth_token", "auth_token_update_strategy") {
			params := &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:        aws.Bool(true),
				ReplicationGroupId:      aws.String(d.Id()),
				AuthTokenUpdateStrategy: aws.String(d.Get("auth_token_update_strategy").(string)),
				AuthToken:               aws.String(d.Get("auth_token").(string)),
			}

			_, err := conn.ModifyReplicationGroupWithContext(ctx, params)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "changing auth_token for ElastiCache Replication Group (%s): %s", d.Id(), err)
			}

			_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) auth_token change: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceReplicationGroupRead(ctx, d, meta)...)
}

func resourceReplicationGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	v, hasGlobalReplicationGroupID := d.GetOk("global_replication_group_id")
	if hasGlobalReplicationGroupID {
		globalReplicationGroupID := v.(string)
		err := DisassociateReplicationGroup(ctx, conn, globalReplicationGroupID, d.Id(), meta.(*conns.AWSClient).Region, GlobalReplicationGroupDisassociationReadyTimeout)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disassociating ElastiCache Replication Group (%s) from Global Replication Group (%s): %s", d.Id(), globalReplicationGroupID, err)
		}
	}

	var finalSnapshotID = d.Get("final_snapshot_identifier").(string)
	err := deleteReplicationGroup(ctx, d.Id(), conn, finalSnapshotID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache Replication Group (%s): %s", d.Id(), err)
	}

	if hasGlobalReplicationGroupID {
		paramGroupName := d.Get("parameter_group_name").(string)
		if paramGroupName != "" {
			err := deleteParameterGroup(ctx, conn, paramGroupName)
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
				return diags
			}
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting ElastiCache Parameter Group (%s): %s", d.Id(), err)
			}
		}
	}

	return diags
}

func DisassociateReplicationGroup(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID, id, region string, readyTimeout time.Duration) error {
	input := &elasticache.DisassociateGlobalReplicationGroupInput{
		GlobalReplicationGroupId: aws.String(globalReplicationGroupID),
		ReplicationGroupId:       aws.String(id),
		ReplicationGroupRegion:   aws.String(region),
	}
	err := retry.RetryContext(ctx, readyTimeout, func() *retry.RetryError {
		_, err := conn.DisassociateGlobalReplicationGroupWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DisassociateGlobalReplicationGroupWithContext(ctx, input)
	}
	if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "is not associated with Global Replication Group") {
		return nil
	}
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault) {
		return fmt.Errorf("tried for %s: %w", readyTimeout.String(), err)
	}

	if err != nil {
		return err
	}

	_, err = waitGlobalReplicationGroupMemberDetached(ctx, conn, globalReplicationGroupID, id)
	if err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func deleteReplicationGroup(ctx context.Context, replicationGroupID string, conn *elasticache.ElastiCache, finalSnapshotID string, timeout time.Duration) error {
	input := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	if finalSnapshotID != "" {
		input.FinalSnapshotIdentifier = aws.String(finalSnapshotID)
	}

	// 10 minutes should give any creating/deleting cache clusters or snapshots time to complete
	err := retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteReplicationGroupWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
			return nil
		}
		// Cache Cluster is creating/deleting or Replication Group is snapshotting
		// InvalidReplicationGroupState: Cache cluster tf-acc-test-uqhe-003 is not in a valid state to be deleted
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidReplicationGroupStateFault) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteReplicationGroupWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = WaitReplicationGroupDeleted(ctx, conn, replicationGroupID, timeout)

	return err
}

func modifyReplicationGroupShardConfiguration(ctx context.Context, conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	if d.HasChange("num_node_groups") {
		err := modifyReplicationGroupShardConfigurationNumNodeGroups(ctx, conn, d, "num_node_groups")
		if err != nil {
			return err
		}
	}

	if d.HasChange("replicas_per_node_group") {
		err := modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(ctx, conn, d, "replicas_per_node_group")
		if err != nil {
			return err
		}
	}

	return nil
}

func modifyReplicationGroupShardConfigurationNumNodeGroups(ctx context.Context, conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldNumNodeGroups := o.(int)
	newNumNodeGroups := n.(int)

	input := &elasticache.ModifyReplicationGroupShardConfigurationInput{
		ApplyImmediately:   aws.Bool(true),
		NodeGroupCount:     aws.Int64(int64(newNumNodeGroups)),
		ReplicationGroupId: aws.String(d.Id()),
	}

	if oldNumNodeGroups > newNumNodeGroups {
		// Node Group IDs are 1 indexed: 0001 through 0015
		// Loop from highest old ID until we reach highest new ID
		nodeGroupsToRemove := []string{}
		for i := oldNumNodeGroups; i > newNumNodeGroups; i-- {
			nodeGroupID := fmt.Sprintf("%04d", i)
			nodeGroupsToRemove = append(nodeGroupsToRemove, nodeGroupID)
		}
		input.NodeGroupsToRemove = aws.StringSlice(nodeGroupsToRemove)
	}

	log.Printf("[DEBUG] Modifying ElastiCache Replication Group (%s) shard configuration: %s", d.Id(), input)
	_, err := conn.ModifyReplicationGroupShardConfigurationWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("modifying ElastiCache Replication Group shard configuration: %w", err)
	}

	_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) shard reconfiguration completion: %w", d.Id(), err)
	}

	return nil
}

func modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(ctx context.Context, conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldReplicas := o.(int)
	newReplicas := n.(int)

	if newReplicas > oldReplicas {
		input := &elasticache.IncreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.IncreaseReplicaCountWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("adding ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("waiting for ElastiCache Replication Group (%s) replica addition: %w", d.Id(), err)
		}
	} else {
		input := &elasticache.DecreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.DecreaseReplicaCountWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("removing ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = WaitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("waiting for ElastiCache Replication Group (%s) replica removal: %w", d.Id(), err)
		}
	}

	return nil
}

func modifyReplicationGroupNumCacheClusters(ctx context.Context, conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldNumberCacheClusters := o.(int)
	newNumberCacheClusters := n.(int)

	var err error
	if newNumberCacheClusters > oldNumberCacheClusters {
		err = increaseReplicationGroupNumCacheClusters(ctx, conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	} else if newNumberCacheClusters < oldNumberCacheClusters {
		err = decreaseReplicationGroupNumCacheClusters(ctx, conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	}
	return err
}

func increaseReplicationGroupNumCacheClusters(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.IncreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.IncreaseReplicaCountWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("adding ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = WaitReplicationGroupMemberClustersAvailable(ctx, conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) replica addition: %w", replicationGroupID, err)
	}

	return nil
}

func decreaseReplicationGroupNumCacheClusters(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.DecreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.DecreaseReplicaCountWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("removing ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = WaitReplicationGroupMemberClustersAvailable(ctx, conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) replica removal: %w", replicationGroupID, err)
	}

	return nil
}

var validateReplicationGroupID schema.SchemaValidateFunc = validation.All(
	validation.StringLenBetween(1, 40),
	validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric characters and hyphens"),
	validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with a letter"),
	validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
	validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
)
