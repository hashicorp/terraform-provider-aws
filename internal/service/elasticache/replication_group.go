// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	failoverMinNumCacheClusters = 2
)

// @SDKResource("aws_elasticache_replication_group", name="Replication Group")
// @Tags(identifierAttribute="arn")
func resourceReplicationGroup() *schema.Resource {
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
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthTokenUpdateStrategyType](),
				Default:          awstypes.AuthTokenUpdateStrategyTypeRotate,
			},
			names.AttrAutoMinorVersionUpgrade: {
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
			"cluster_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ClusterMode](),
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
			names.AttrDescription: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			names.AttrEngine: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineRedis,
				ValidateFunc: validation.StringInSlice([]string{engineRedis}, true),
			},
			names.AttrEngineVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validRedisVersionString,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ConflictsWith: []string{
					"num_node_groups",
					names.AttrParameterGroupName,
					names.AttrEngine,
					names.AttrEngineVersion,
					"node_type",
					"security_group_names",
					"transit_encryption_enabled",
					"transit_encryption_mode",
					"at_rest_encryption_enabled",
					"snapshot_arns",
					"snapshot_name",
				},
			},
			"ip_discovery": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpDiscovery](),
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"log_delivery_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DestinationType](),
						},
						names.AttrDestination: {
							Type:     schema.TypeString,
							Required: true,
						},
						"log_format": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogFormat](),
						},
						"log_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogType](),
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
			},
			"multi_az_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"network_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.NetworkType](),
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
				ConflictsWith: []string{"num_node_groups", "replicas_per_node_group"},
			},
			"num_node_groups": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"num_cache_clusters", "global_replication_group_id"},
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.HasPrefix(old, "global-datastore-")
				},
			},
			names.AttrPort: {
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
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"num_cache_clusters"},
				ValidateFunc:  validation.IntBetween(0, 5),
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
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
				Computed: true,
			},
			"transit_encryption_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TransitEncryptionMode](),
			},
			"user_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"auth_token"},
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
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			replicationGroupValidateMultiAZAutomaticFailover,
			customizeDiffEngineVersionForceNewOnDowngrade,
			customdiff.ComputedIf("member_clusters", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("num_cache_clusters") ||
					diff.HasChange("num_node_groups") ||
					diff.HasChange("replicas_per_node_group")
			}),
			customdiff.ForceNewIf("transit_encryption_enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				// For Redis engine versions < 7.0.5, transit_encryption_enabled can only
				// be configured during creation of the cluster.
				return semver.LessThan(d.Get("engine_version_actual").(string), "7.0.5")
			}),
			replicationGroupValidateAutomaticFailoverNumCacheClusters,
			verify.SetTagsDiff,
		),
	}
}

func resourceReplicationGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	replicationGroupID := d.Get("replication_group_id").(string)
	input := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
		Tags:               getTagsIn(ctx),
	}

	if _, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		input.AtRestEncryptionEnabled = aws.Bool(d.Get("at_rest_encryption_enabled").(bool))
	}

	if v, ok := d.GetOk("auth_token"); ok {
		input.AuthToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrAutoMinorVersionUpgrade); ok {
		if v, null, _ := nullable.Bool(v.(string)).ValueBool(); !null {
			input.AutoMinorVersionUpgrade = aws.Bool(v)
		}
	}

	if v, ok := d.GetOk("cluster_mode"); ok {
		input.ClusterMode = awstypes.ClusterMode(v.(string))
	}

	if v, ok := d.GetOk("data_tiering_enabled"); ok {
		input.DataTieringEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ReplicationGroupDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
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
		input.Engine = aws.String(d.Get(names.AttrEngine).(string))
	}

	if v, ok := d.GetOk("ip_discovery"); ok {
		input.IpDiscovery = awstypes.IpDiscovery(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_delivery_configuration"); ok && v.(*schema.Set).Len() > 0 {
		for _, tfMapRaw := range v.(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			apiObject := expandLogDeliveryConfigurationRequests(tfMap)
			input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, apiObject)
		}
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_az_enabled"); ok {
		input.MultiAZEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("network_type"); ok {
		input.NetworkType = awstypes.NetworkType(v.(string))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		input.NotificationTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("num_cache_clusters"); ok {
		input.NumCacheClusters = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("num_node_groups"); ok && v != 0 {
		input.NumNodeGroups = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
		input.CacheParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("preferred_cache_cluster_azs"); ok && len(v.([]interface{})) > 0 {
		input.PreferredCacheClusterAZs = flex.ExpandStringValueList(v.([]interface{}))
	}

	rawConfig := d.GetRawConfig()
	rawReplicasPerNodeGroup := rawConfig.GetAttr("replicas_per_node_group")
	if rawReplicasPerNodeGroup.IsKnown() && !rawReplicasPerNodeGroup.IsNull() {
		var v int32
		err := gocty.FromCtyValue(rawReplicasPerNodeGroup, &v)
		if err != nil {
			path := cty.GetAttrPath("replicas_per_node_group")
			diags = append(diags, errs.NewAttributeErrorDiagnostic(
				path,
				"Invalid Value",
				"An unexpected error occurred while reading configuration values. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					fmt.Sprintf(`Reading "%s": %s`, errs.PathString(path), err),
			))
		}
		input.ReplicasPerNodeGroup = aws.Int32(v)
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		input.CacheSubnetGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_group_names"); ok && v.(*schema.Set).Len() > 0 {
		input.CacheSecurityGroupNames = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("snapshot_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.SnapshotArns = flex.ExpandStringValueSet(v.(*schema.Set))
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

	if v, ok := d.GetOk("transit_encryption_enabled"); ok {
		input.TransitEncryptionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("transit_encryption_mode"); ok {
		input.TransitEncryptionMode = awstypes.TransitEncryptionMode(v.(string))
	}

	if v, ok := d.GetOk("user_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.UserGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateReplicationGroup(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateReplicationGroup(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache Replication Group (%s): %s", replicationGroupID, err)
	}

	d.SetId(aws.ToString(output.ReplicationGroup.ReplicationGroupId))

	const (
		delay = 30 * time.Second
	)
	if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate), delay); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		// When adding a replication group to a global replication group, the replication group can be in the "available"
		// state, but the global replication group can still be in the "modifying" state. Wait for the replication group
		// to be fully added to the global replication group.
		// API calls to the global replication group can be made in any region.
		if _, err := waitGlobalReplicationGroupAvailable(ctx, conn, v.(string), globalReplicationGroupDefaultCreatedTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Global Replication Group (%s) available: %s", v, err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.ToString(output.ReplicationGroup.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	rgp, err := findReplicationGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): %s", d.Id(), err)
	}

	if aws.ToString(rgp.Status) == replicationGroupStatusDeleting {
		log.Printf("[WARN] ElastiCache Replication Group (%s) is currently in the `deleting` status, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if rgp.GlobalReplicationGroupInfo != nil && rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId != nil {
		d.Set("global_replication_group_id", rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
	}

	switch rgp.AutomaticFailover {
	case awstypes.AutomaticFailoverStatusDisabled, awstypes.AutomaticFailoverStatusDisabling:
		d.Set("automatic_failover_enabled", false)
	case awstypes.AutomaticFailoverStatusEnabled, awstypes.AutomaticFailoverStatusEnabling:
		d.Set("automatic_failover_enabled", true)
	default:
		log.Printf("Unknown AutomaticFailover state %q", string(rgp.AutomaticFailover))
	}

	switch rgp.MultiAZ {
	case awstypes.MultiAZStatusEnabled:
		d.Set("multi_az_enabled", true)
	case awstypes.MultiAZStatusDisabled:
		d.Set("multi_az_enabled", false)
	default:
		log.Printf("Unknown MultiAZ state %q", string(rgp.MultiAZ))
	}

	d.Set(names.AttrKMSKeyID, rgp.KmsKeyId)
	d.Set(names.AttrDescription, rgp.Description)
	d.Set("num_cache_clusters", len(rgp.MemberClusters))
	if err := d.Set("member_clusters", flex.FlattenStringValueSet(rgp.MemberClusters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting member_clusters: %s", err)
	}

	d.Set("num_node_groups", len(rgp.NodeGroups))
	d.Set("replicas_per_node_group", len(rgp.NodeGroups[0].NodeGroupMembers)-1)

	d.Set("cluster_enabled", rgp.ClusterEnabled)
	d.Set("cluster_mode", rgp.ClusterMode)
	d.Set("replication_group_id", rgp.ReplicationGroupId)
	d.Set(names.AttrARN, rgp.ARN)
	d.Set("data_tiering_enabled", rgp.DataTiering == awstypes.DataTieringStatusEnabled)

	d.Set("ip_discovery", rgp.IpDiscovery)
	d.Set("network_type", rgp.NetworkType)

	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(rgp.LogDeliveryConfigurations))
	d.Set("snapshot_window", rgp.SnapshotWindow)
	d.Set("snapshot_retention_limit", rgp.SnapshotRetentionLimit)

	if rgp.ConfigurationEndpoint != nil {
		d.Set(names.AttrPort, rgp.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint_address", rgp.ConfigurationEndpoint.Address)
	} else {
		log.Printf("[DEBUG] ElastiCache Replication Group (%s) Configuration Endpoint is nil", d.Id())

		if rgp.NodeGroups[0].PrimaryEndpoint != nil {
			log.Printf("[DEBUG] ElastiCache Replication Group (%s) Primary Endpoint is not nil", d.Id())
			d.Set(names.AttrPort, rgp.NodeGroups[0].PrimaryEndpoint.Port)
			d.Set("primary_endpoint_address", rgp.NodeGroups[0].PrimaryEndpoint.Address)
		}

		if rgp.NodeGroups[0].ReaderEndpoint != nil {
			d.Set("reader_endpoint_address", rgp.NodeGroups[0].ReaderEndpoint.Address)
		}
	}

	d.Set("user_group_ids", rgp.UserGroupIds)

	// Tags cannot be read when the replication group is not Available
	log.Printf("[DEBUG] Waiting for ElastiCache Replication Group (%s) to become available", d.Id())

	const (
		delay = 0 * time.Second
	)
	if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) create: %s", aws.ToString(rgp.ARN), err)
	}

	log.Printf("[DEBUG] ElastiCache Replication Group (%s): Checking underlying cache clusters", d.Id())

	// This section reads settings that require checking the underlying cache clusters
	if rgp.NodeGroups != nil && len(rgp.NodeGroups[0].NodeGroupMembers) != 0 {
		cacheCluster := rgp.NodeGroups[0].NodeGroupMembers[0]
		input := &elasticache.DescribeCacheClustersInput{
			CacheClusterId:    cacheCluster.CacheClusterId,
			ShowCacheNodeInfo: aws.Bool(true),
		}

		output, err := conn.DescribeCacheClusters(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): reading Cache Cluster (%s): %s", d.Id(), aws.ToString(cacheCluster.CacheClusterId), err)
		}

		if len(output.CacheClusters) == 0 {
			return diags
		}

		c := output.CacheClusters[0]

		if err := setFromCacheCluster(d, &c); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): reading Cache Cluster (%s): %s", d.Id(), aws.ToString(cacheCluster.CacheClusterId), err)
		}

		d.Set("at_rest_encryption_enabled", c.AtRestEncryptionEnabled)
		d.Set("transit_encryption_enabled", c.TransitEncryptionEnabled)
		d.Set("transit_encryption_mode", c.TransitEncryptionMode)

		if c.AuthTokenEnabled != nil && !aws.ToBool(c.AuthTokenEnabled) {
			d.Set("auth_token", nil)
		}
	}

	return diags
}

func resourceReplicationGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		o, n := d.GetChange("num_cache_clusters")
		oldCacheClusterCount, newCacheClusterCount := o.(int), n.(int)

		if d.HasChanges("num_node_groups", "replicas_per_node_group") {
			if err := modifyReplicationGroupShardConfiguration(ctx, conn, d); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else if d.HasChange("num_cache_clusters") {
			if newCacheClusterCount > oldCacheClusterCount {
				if err := increaseReplicationGroupReplicaCount(ctx, conn, d.Id(), newCacheClusterCount, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			} // Else defer until after all other modifications are made.
		}

		requestUpdate := false
		input := &elasticache.ModifyReplicationGroupInput{
			ApplyImmediately:   aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			ReplicationGroupId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAutoMinorVersionUpgrade) {
			if v, ok := d.GetOk(names.AttrAutoMinorVersionUpgrade); ok {
				if v, null, _ := nullable.Bool(v.(string)).ValueBool(); !null {
					input.AutoMinorVersionUpgrade = aws.Bool(v)
					requestUpdate = true
				}
			}
		}

		if d.HasChange("automatic_failover_enabled") {
			input.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
			requestUpdate = true
		}

		if d.HasChange(names.AttrDescription) {
			input.ReplicationGroupDescription = aws.String(d.Get(names.AttrDescription).(string))
			requestUpdate = true
		}

		if d.HasChange("cluster_mode") {
			input.ClusterMode = awstypes.ClusterMode(d.Get("cluster_mode").(string))
			requestUpdate = true
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
			requestUpdate = true
		}

		if d.HasChange("ip_discovery") {
			input.IpDiscovery = awstypes.IpDiscovery(d.Get("ip_discovery").(string))
			requestUpdate = true
		}

		if d.HasChange("log_delivery_configuration") {
			o, n := d.GetChange("log_delivery_configuration")

			input.LogDeliveryConfigurations = []awstypes.LogDeliveryConfigurationRequest{}
			logTypesToSubmit := make(map[awstypes.LogType]bool)

			currentLogDeliveryConfig := n.(*schema.Set).List()
			for _, current := range currentLogDeliveryConfig {
				logDeliveryConfigurationRequest := expandLogDeliveryConfigurationRequests(current.(map[string]interface{}))
				logTypesToSubmit[logDeliveryConfigurationRequest.LogType] = true
				input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, logDeliveryConfigurationRequest)
			}

			previousLogDeliveryConfig := o.(*schema.Set).List()
			for _, previous := range previousLogDeliveryConfig {
				logDeliveryConfigurationRequest := expandEmptyLogDeliveryConfigurationRequest(previous.(map[string]interface{}))
				//if something was removed, send an empty request
				if !logTypesToSubmit[logDeliveryConfigurationRequest.LogType] {
					input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, logDeliveryConfigurationRequest)
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

		if d.HasChange("network_type") {
			input.IpDiscovery = awstypes.IpDiscovery(d.Get("network_type").(string))
			requestUpdate = true
		}

		if d.HasChange("node_type") {
			input.CacheNodeType = aws.String(d.Get("node_type").(string))
			requestUpdate = true
		}

		if d.HasChange("notification_topic_arn") {
			input.NotificationTopicArn = aws.String(d.Get("notification_topic_arn").(string))
			requestUpdate = true
		}

		if d.HasChange(names.AttrParameterGroupName) {
			input.CacheParameterGroupName = aws.String(d.Get(names.AttrParameterGroupName).(string))
			requestUpdate = true
		}

		if d.HasChange(names.AttrSecurityGroupIDs) {
			if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
				input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
				requestUpdate = true
			}
		}

		if d.HasChange("security_group_names") {
			if v, ok := d.GetOk("security_group_names"); ok && v.(*schema.Set).Len() > 0 {
				input.CacheSecurityGroupNames = flex.ExpandStringValueSet(v.(*schema.Set))
				requestUpdate = true
			}
		}

		if d.HasChange("snapshot_retention_limit") {
			// This is a real hack to set the Snapshotting Cluster ID to be the first Cluster in the RG.
			o, _ := d.GetChange("snapshot_retention_limit")
			if o.(int) == 0 {
				input.SnapshottingClusterId = aws.String(fmt.Sprintf("%s-001", d.Id()))
			}

			input.SnapshotRetentionLimit = aws.Int32(int32(d.Get("snapshot_retention_limit").(int)))
			requestUpdate = true
		}

		if d.HasChange("snapshot_window") {
			input.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
			requestUpdate = true
		}

		if d.HasChange("transit_encryption_enabled") {
			input.TransitEncryptionEnabled = aws.Bool(d.Get("transit_encryption_enabled").(bool))
			requestUpdate = true
		}

		if d.HasChange("transit_encryption_mode") {
			input.TransitEncryptionMode = awstypes.TransitEncryptionMode(d.Get("transit_encryption_mode").(string))
			requestUpdate = true
		}

		if d.HasChange("user_group_ids") {
			o, n := d.GetChange("user_group_ids")
			ns, os := n.(*schema.Set), o.(*schema.Set)
			add, del := ns.Difference(os), os.Difference(ns)

			if add.Len() > 0 {
				input.UserGroupIdsToAdd = flex.ExpandStringValueSet(add)
				requestUpdate = true
			}

			if del.Len() > 0 {
				input.UserGroupIdsToRemove = flex.ExpandStringValueSet(del)
				requestUpdate = true
			}
		}

		if requestUpdate {
			// tagging may cause this resource to not yet be available, so wait for it to be available
			const (
				delay = 30 * time.Second
			)
			if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) update: %s", d.Id(), err)
			}

			_, err := conn.ModifyReplicationGroup(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ElastiCache Replication Group (%s): %s", d.Id(), err)
			}

			if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) update: %s", d.Id(), err)
			}
		}

		if d.HasChanges("auth_token", "auth_token_update_strategy") {
			input := &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:        aws.Bool(true),
				AuthToken:               aws.String(d.Get("auth_token").(string)),
				AuthTokenUpdateStrategy: awstypes.AuthTokenUpdateStrategyType(d.Get("auth_token_update_strategy").(string)),
				ReplicationGroupId:      aws.String(d.Id()),
			}

			// tagging may cause this resource to not yet be available, so wait for it to be available
			const (
				delay = 0 * time.Second
			)
			if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) update: %s", d.Id(), err)
			}

			_, err := conn.ModifyReplicationGroup(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ElastiCache Replication Group (%s) authentication: %s", d.Id(), err)
			}

			if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) update: %s", d.Id(), err)
			}
		}

		if d.HasChange("num_cache_clusters") {
			if newCacheClusterCount < oldCacheClusterCount {
				if err := decreaseReplicationGroupReplicaCount(ctx, conn, d.Id(), newCacheClusterCount, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceReplicationGroupRead(ctx, d, meta)...)
}

func resourceReplicationGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	v, hasGlobalReplicationGroupID := d.GetOk("global_replication_group_id")
	if hasGlobalReplicationGroupID {
		if err := disassociateReplicationGroup(ctx, conn, v.(string), d.Id(), meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	input := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
		input.FinalSnapshotIdentifier = aws.String(v.(string))
	}

	// Cache Cluster is creating/deleting or Replication Group is snapshotting
	// InvalidReplicationGroupState: Cache cluster tf-acc-test-uqhe-003 is not in a valid state to be deleted
	const (
		timeout = 10 * time.Minute // 10 minutes should give any creating/deleting cache clusters or snapshots time to complete.
	)
	log.Printf("[INFO] Deleting ElastiCache Replication Group: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidReplicationGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteReplicationGroup(ctx, input)
	})

	switch {
	case errs.IsA[*awstypes.ReplicationGroupNotFoundFault](err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache Replication Group (%s): %s", d.Id(), err)
	default:
		if _, err := waitReplicationGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Replication Group (%s) delete: %s", d.Id(), err)
		}
	}

	if hasGlobalReplicationGroupID {
		if paramGroupName := d.Get(names.AttrParameterGroupName).(string); paramGroupName != "" {
			if err := deleteParameterGroup(ctx, conn, paramGroupName); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return diags
}

func disassociateReplicationGroup(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID, replicationGroupID, region string, timeout time.Duration) error {
	input := &elasticache.DisassociateGlobalReplicationGroupInput{
		GlobalReplicationGroupId: aws.String(globalReplicationGroupID),
		ReplicationGroupId:       aws.String(replicationGroupID),
		ReplicationGroupRegion:   aws.String(region),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidGlobalReplicationGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.DisassociateGlobalReplicationGroup(ctx, input)
	})

	if errs.IsA[*awstypes.GlobalReplicationGroupNotFoundFault](err) {
		return nil
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "is not associated with Global Replication Group") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("disassociating ElastiCache Replication Group (%s) from Global Replication Group (%s): %w", replicationGroupID, globalReplicationGroupID, err)
	}

	if _, err := waitGlobalReplicationGroupMemberDetached(ctx, conn, globalReplicationGroupID, replicationGroupID, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) detach: %w", replicationGroupID, err)
	}

	return nil
}

func modifyReplicationGroupShardConfiguration(ctx context.Context, conn *elasticache.Client, d *schema.ResourceData) error {
	if d.HasChange("num_node_groups") {
		if err := modifyReplicationGroupShardConfigurationNumNodeGroups(ctx, conn, d, "num_node_groups"); err != nil {
			return err
		}
	}

	if d.HasChange("replicas_per_node_group") {
		if err := modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(ctx, conn, d, "replicas_per_node_group"); err != nil {
			return err
		}
	}

	return nil
}

func modifyReplicationGroupShardConfigurationNumNodeGroups(ctx context.Context, conn *elasticache.Client, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldNodeGroupCount, newNodeGroupCount := o.(int), n.(int)

	input := &elasticache.ModifyReplicationGroupShardConfigurationInput{
		ApplyImmediately:   aws.Bool(true),
		NodeGroupCount:     aws.Int32(int32(newNodeGroupCount)),
		ReplicationGroupId: aws.String(d.Id()),
	}

	if oldNodeGroupCount > newNodeGroupCount {
		// Node Group IDs are 1 indexed: 0001 through 0015
		// Loop from highest old ID until we reach highest new ID
		nodeGroupsToRemove := []string{}
		for i := oldNodeGroupCount; i > newNodeGroupCount; i-- {
			nodeGroupID := fmt.Sprintf("%04d", i)
			nodeGroupsToRemove = append(nodeGroupsToRemove, nodeGroupID)
		}
		input.NodeGroupsToRemove = nodeGroupsToRemove
	}

	_, err := conn.ModifyReplicationGroupShardConfiguration(ctx, input)

	if err != nil {
		return fmt.Errorf("modifying ElastiCache Replication Group (%s) shard configuration: %w", d.Id(), err)
	}

	const (
		delay = 30 * time.Second
	)
	if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) update: %w", d.Id(), err)
	}

	return nil
}

func modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(ctx context.Context, conn *elasticache.Client, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldReplicaCount, newReplicaCount := o.(int), n.(int)

	if newReplicaCount > oldReplicaCount {
		input := &elasticache.IncreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int32(int32(newReplicaCount)),
			ReplicationGroupId: aws.String(d.Id()),
		}

		_, err := conn.IncreaseReplicaCount(ctx, input)

		if err != nil {
			return fmt.Errorf("increasing ElastiCache Replication Group (%s) replica count (%d): %w", d.Id(), newReplicaCount, err)
		}

		const (
			delay = 30 * time.Second
		)
		if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
			return fmt.Errorf("waiting for ElastiCache Replication Group (%s) update: %w", d.Id(), err)
		}
	} else if newReplicaCount < oldReplicaCount {
		input := &elasticache.DecreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int32(int32(newReplicaCount)),
			ReplicationGroupId: aws.String(d.Id()),
		}

		_, err := conn.DecreaseReplicaCount(ctx, input)

		if err != nil {
			return fmt.Errorf("decreasing ElastiCache Replication Group (%s) replica count (%d): %w", d.Id(), newReplicaCount, err)
		}

		const (
			delay = 30 * time.Second
		)
		if _, err := waitReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), delay); err != nil {
			return fmt.Errorf("waiting for ElastiCache Replication Group (%s) update: %w", d.Id(), err)
		}
	}

	return nil
}

func increaseReplicationGroupReplicaCount(ctx context.Context, conn *elasticache.Client, replicationGroupID string, newReplicaCount int, timeout time.Duration) error {
	input := &elasticache.IncreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int32(int32(newReplicaCount - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}

	_, err := conn.IncreaseReplicaCount(ctx, input)

	if err != nil {
		return fmt.Errorf("increasing ElastiCache Replication Group (%s) replica count (%d): %w", replicationGroupID, newReplicaCount-1, err)
	}

	if _, err := waitReplicationGroupMemberClustersAvailable(ctx, conn, replicationGroupID, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) member cluster update: %w", replicationGroupID, err)
	}

	return nil
}

func decreaseReplicationGroupReplicaCount(ctx context.Context, conn *elasticache.Client, replicationGroupID string, newReplicaCount int, timeout time.Duration) error {
	input := &elasticache.DecreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int32(int32(newReplicaCount - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}

	_, err := conn.DecreaseReplicaCount(ctx, input)

	if err != nil {
		return fmt.Errorf("decreasing ElastiCache Replication Group (%s) replica count (%d): %w", replicationGroupID, newReplicaCount-1, err)
	}

	if _, err := waitReplicationGroupMemberClustersAvailable(ctx, conn, replicationGroupID, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group (%s) member cluster update: %w", replicationGroupID, err)
	}

	return nil
}

func findReplicationGroupByID(ctx context.Context, conn *elasticache.Client, id string) (*awstypes.ReplicationGroup, error) {
	input := &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(id),
	}

	return findReplicationGroup(ctx, conn, input, tfslices.PredicateTrue[*awstypes.ReplicationGroup]())
}

func findReplicationGroup(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeReplicationGroupsInput, filter tfslices.Predicate[*awstypes.ReplicationGroup]) (*awstypes.ReplicationGroup, error) {
	output, err := findReplicationGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationGroups(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeReplicationGroupsInput, filter tfslices.Predicate[*awstypes.ReplicationGroup]) ([]awstypes.ReplicationGroup, error) {
	var output []awstypes.ReplicationGroup

	pages := elasticache.NewDescribeReplicationGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ReplicationGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReplicationGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusReplicationGroup(ctx context.Context, conn *elasticache.Client, replicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationGroupByID(ctx, conn, replicationGroupID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

const (
	replicationGroupStatusAvailable    = "available"
	replicationGroupStatusCreateFailed = "create-failed"
	replicationGroupStatusCreating     = "creating"
	replicationGroupStatusDeleting     = "deleting"
	replicationGroupStatusModifying    = "modifying"
	replicationGroupStatusSnapshotting = "snapshotting"
)

func waitReplicationGroupAvailable(ctx context.Context, conn *elasticache.Client, replicationGroupID string, timeout time.Duration, delay time.Duration) (*awstypes.ReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			replicationGroupStatusCreating,
			replicationGroupStatusModifying,
			replicationGroupStatusSnapshotting,
		},
		Target:     []string{replicationGroupStatusAvailable},
		Refresh:    statusReplicationGroup(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      delay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationGroup); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationGroupDeleted(ctx context.Context, conn *elasticache.Client, replicationGroupID string, timeout time.Duration) (*awstypes.ReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			replicationGroupStatusCreating,
			replicationGroupStatusAvailable,
			replicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    statusReplicationGroup(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationGroup); ok {
		return output, err
	}

	return nil, err
}

func findReplicationGroupMemberClustersByID(ctx context.Context, conn *elasticache.Client, id string) ([]awstypes.CacheCluster, error) {
	rg, err := findReplicationGroupByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}
	ids := rg.MemberClusters
	clusters, err := findCacheClusters(ctx, conn, &elasticache.DescribeCacheClustersInput{}, func(v *awstypes.CacheCluster) bool {
		return slices.Contains(ids, aws.ToString(v.CacheClusterId))
	})

	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return clusters, nil
}

// statusReplicationGroupMemberClusters fetches the Replication Group's Member Clusters and either "available" or the first non-"available" status.
// NOTE: This function assumes that the intended end-state is to have all member clusters in "available" status.
func statusReplicationGroupMemberClusters(ctx context.Context, conn *elasticache.Client, replicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationGroupMemberClustersByID(ctx, conn, replicationGroupID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := cacheClusterStatusAvailable
		for _, v := range output {
			if clusterStatus := aws.ToString(v.CacheClusterStatus); clusterStatus != cacheClusterStatusAvailable {
				status = clusterStatus
				break
			}
		}

		return output, status, nil
	}
}

func waitReplicationGroupMemberClustersAvailable(ctx context.Context, conn *elasticache.Client, replicationGroupID string, timeout time.Duration) ([]*awstypes.CacheCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			cacheClusterStatusCreating,
			cacheClusterStatusDeleting,
			cacheClusterStatusModifying,
			cacheClusterStatusSnapshotting,
		},
		Target:     []string{cacheClusterStatusAvailable},
		Refresh:    statusReplicationGroupMemberClusters(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.CacheCluster); ok {
		return output, err
	}

	return nil, err
}

var validateReplicationGroupID schema.SchemaValidateFunc = validation.All(
	validation.StringLenBetween(1, 40),
	validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric characters and hyphens"),
	validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with a letter"),
	validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
	validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
)

// replicationGroupValidateMultiAZAutomaticFailover validates that `automatic_failover_enabled` is set when `multi_az_enabled` is true
func replicationGroupValidateMultiAZAutomaticFailover(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v := diff.Get("multi_az_enabled").(bool); !v {
		return nil
	}
	if v := diff.Get("automatic_failover_enabled").(bool); !v {
		return errors.New(`automatic_failover_enabled must be true if multi_az_enabled is true`)
	}
	return nil
}

// replicationGroupValidateAutomaticFailoverNumCacheClusters validates that `automatic_failover_enabled` is set when `multi_az_enabled` is true
func replicationGroupValidateAutomaticFailoverNumCacheClusters(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v := diff.Get("automatic_failover_enabled").(bool); !v {
		return nil
	}
	raw := diff.GetRawConfig().GetAttr("num_cache_clusters")
	if !raw.IsKnown() || raw.IsNull() {
		return nil
	}
	if raw.GreaterThanOrEqualTo(cty.NumberIntVal(failoverMinNumCacheClusters)).True() {
		return nil
	}
	return errors.New(`"num_cache_clusters": must be at least 2 if automatic_failover_enabled is true`)
}
