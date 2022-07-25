package elasticache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicationGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceReplicationGroupCreate,
		Read:   resourceReplicationGroupRead,
		Update: resourceReplicationGroupUpdate,
		Delete: resourceReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"availability_zones": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				ConflictsWith: []string{"preferred_cache_cluster_azs"},
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_mode": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"num_node_groups", "replicas_per_node_group"},
				Deprecated:    "Use num_node_groups and replicas_per_node_group instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"num_node_groups": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"num_node_groups", "number_cache_clusters", "num_cache_clusters", "global_replication_group_id"},
							Deprecated:    "Use root-level num_node_groups instead",
						},
						"replicas_per_node_group": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"replicas_per_node_group"},
							Deprecated:    "Use root-level replicas_per_node_group instead",
						},
					},
				},
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
				ExactlyOneOf: []string{"description", "replication_group_description"},
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
					"cluster_mode.0.num_node_groups",
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
				ConflictsWith: []string{"cluster_mode.0.num_node_groups", "num_node_groups", "number_cache_clusters"},
			},
			"num_node_groups": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"cluster_mode", "number_cache_clusters", "num_cache_clusters", "global_replication_group_id"},
			},
			"number_cache_clusters": {
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"cluster_mode.0.num_node_groups", "num_cache_clusters", "num_node_groups"},
				Deprecated:    "Use num_cache_clusters instead",
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
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"availability_zones"},
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
				ConflictsWith: []string{"cluster_mode"},
			},
			"replication_group_description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"description", "replication_group_description"},
				Deprecated:   "Use description instead",
				ValidateFunc: validation.StringIsNotEmpty,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
		SchemaVersion: 1,

		// SchemaVersion: 1 did not include any state changes via MigrateState.
		// Perform a no-operation state upgrade for Terraform 0.12 compatibility.
		// Future state migrations should be performed with StateUpgraders.
		MigrateState: func(v int, inst *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
			return inst, nil
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
				return diff.HasChange("number_cache_clusters") ||
					diff.HasChange("num_cache_clusters") ||
					diff.HasChange("cluster_mode.0.num_node_groups") ||
					diff.HasChange("cluster_mode.0.replicas_per_node_group") ||
					diff.HasChange("num_node_groups") ||
					diff.HasChange("replicas_per_node_group")
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	params := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId: aws.String(d.Get("replication_group_id").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("description"); ok {
		params.ReplicationGroupDescription = aws.String(v.(string))
	}
	if v, ok := d.GetOk("replication_group_description"); ok {
		params.ReplicationGroupDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_tiering_enabled"); ok {
		params.DataTieringEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		params.GlobalReplicationGroupId = aws.String(v.(string))
	} else {
		// This cannot be handled at plan-time
		nodeType := d.Get("node_type").(string)
		if nodeType == "" {
			return errors.New(`"node_type" is required unless "global_replication_group_id" is set.`)
		}
		params.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		params.CacheNodeType = aws.String(nodeType)
		params.Engine = aws.String(d.Get("engine").(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		params.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_minor_version_upgrade"); ok {
		if v, null, _ := nullable.Bool(v.(string)).Value(); !null {
			params.AutoMinorVersionUpgrade = aws.Bool(v)
		}
	}

	if preferredAZs, ok := d.GetOk("preferred_cache_cluster_azs"); ok {
		params.PreferredCacheClusterAZs = flex.ExpandStringList(preferredAZs.([]interface{}))
	}
	if availabilityZones := d.Get("availability_zones").(*schema.Set); availabilityZones.Len() > 0 {
		params.PreferredCacheClusterAZs = flex.ExpandStringSet(availabilityZones)
	}

	if v, ok := d.GetOk("parameter_group_name"); ok {
		params.CacheParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		params.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		params.CacheSubnetGroupName = aws.String(v.(string))
	}

	if SGNames := d.Get("security_group_names").(*schema.Set); SGNames.Len() > 0 {
		params.CacheSecurityGroupNames = flex.ExpandStringSet(SGNames)
	}

	if SGIds := d.Get("security_group_ids").(*schema.Set); SGIds.Len() > 0 {
		params.SecurityGroupIds = flex.ExpandStringSet(SGIds)
	}

	if snaps := d.Get("snapshot_arns").(*schema.Set); snaps.Len() > 0 {
		params.SnapshotArns = flex.ExpandStringSet(snaps)
	}

	if v, ok := d.GetOk("log_delivery_configuration"); ok {
		params.LogDeliveryConfigurations = []*elasticache.LogDeliveryConfigurationRequest{}
		v := v.(*schema.Set).List()
		for _, v := range v {
			logDeliveryConfigurationRequest := expandLogDeliveryConfigurations(v.(map[string]interface{}))
			params.LogDeliveryConfigurations = append(params.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
		}
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		params.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if _, ok := d.GetOk("multi_az_enabled"); ok {
		params.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		params.NotificationTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		params.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		params.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		params.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		params.SnapshotName = aws.String(v.(string))
	}

	if _, ok := d.GetOk("transit_encryption_enabled"); ok {
		params.TransitEncryptionEnabled = aws.Bool(d.Get("transit_encryption_enabled").(bool))
	}

	if _, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		params.AtRestEncryptionEnabled = aws.Bool(d.Get("at_rest_encryption_enabled").(bool))
	}

	if v, ok := d.GetOk("auth_token"); ok {
		params.AuthToken = aws.String(v.(string))
	}

	if clusterMode, ok := d.GetOk("cluster_mode"); ok {
		clusterModeList := clusterMode.([]interface{})
		attributes := clusterModeList[0].(map[string]interface{})

		if v, ok := attributes["num_node_groups"]; ok && v != 0 {
			params.NumNodeGroups = aws.Int64(int64(v.(int)))
		}

		if v, ok := attributes["replicas_per_node_group"]; ok {
			params.ReplicasPerNodeGroup = aws.Int64(int64(v.(int)))
		}
	}

	if v, ok := d.GetOk("num_node_groups"); ok && v != 0 {
		params.NumNodeGroups = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("replicas_per_node_group"); ok {
		params.ReplicasPerNodeGroup = aws.Int64(int64(v.(int)))
	}

	if cacheClusters, ok := d.GetOk("number_cache_clusters"); ok {
		params.NumCacheClusters = aws.Int64(int64(cacheClusters.(int)))
	}

	if numCacheClusters, ok := d.GetOk("num_cache_clusters"); ok {
		params.NumCacheClusters = aws.Int64(int64(numCacheClusters.(int)))
	}

	if userGroupIds := d.Get("user_group_ids").(*schema.Set); userGroupIds.Len() > 0 {
		params.UserGroupIds = flex.ExpandStringSet(userGroupIds)
	}

	resp, err := conn.CreateReplicationGroup(params)

	if params.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache Replication Group with tags: %s. Trying create without tags.", err)

		params.Tags = nil
		resp, err = conn.CreateReplicationGroup(params)
	}

	if err != nil {
		return fmt.Errorf("error creating ElastiCache Replication Group (%s): %w", d.Get("replication_group_id").(string), err)
	}

	d.SetId(aws.StringValue(resp.ReplicationGroup.ReplicationGroupId))

	_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Replication Group (%s): waiting for completion: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		// When adding a replication group to a global replication group, the replication group can be in the "available"
		// state, but the global replication group can still be in the "modifying" state. Wait for the replication group
		// to be fully added to the global replication group.
		// API calls to the global replication group can be made in any region.
		if _, err := WaitGlobalReplicationGroupAvailable(conn, v.(string), GlobalReplicationGroupDefaultCreatedTimeout); err != nil {
			return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) to be available: %w", v, err)
		}
	}

	// In some partitions, only post-create tagging supported
	if params.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, aws.StringValue(resp.ReplicationGroup.ARN), nil, tags)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return fmt.Errorf("failed adding tags after create for ElastiCache Replication Group (%s): %w", d.Id(), err)
			}

			log.Printf("[WARN] failed adding tags after create for ElastiCache Replication Group (%s): %s", d.Id(), err)
		}
	}

	return resourceReplicationGroupRead(d, meta)
}

func resourceReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rgp, err := FindReplicationGroupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}

	if aws.StringValue(rgp.Status) == ReplicationGroupStatusDeleting {
		log.Printf("[WARN] ElastiCache Replication Group (%s) is currently in the `deleting` status, removing from state", d.Id())
		d.SetId("")
		return nil
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
	d.Set("replication_group_description", rgp.Description)
	d.Set("number_cache_clusters", len(rgp.MemberClusters))
	d.Set("num_cache_clusters", len(rgp.MemberClusters))
	if err := d.Set("member_clusters", flex.FlattenStringSet(rgp.MemberClusters)); err != nil {
		return fmt.Errorf("error setting member_clusters: %w", err)
	}
	if err := d.Set("cluster_mode", flattenNodeGroupsToClusterMode(rgp.NodeGroups)); err != nil {
		return fmt.Errorf("error setting cluster_mode attribute: %w", err)
	}

	d.Set("num_node_groups", len(rgp.NodeGroups))
	d.Set("replicas_per_node_group", len(rgp.NodeGroups[0].NodeGroupMembers)-1)

	d.Set("cluster_enabled", rgp.ClusterEnabled)
	d.Set("replication_group_id", rgp.ReplicationGroupId)
	d.Set("arn", rgp.ARN)
	d.Set("data_tiering_enabled", aws.StringValue(rgp.DataTiering) == elasticache.DataTieringStatusEnabled)

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
			if rgp.NodeGroups[0].PrimaryEndpoint.Port != nil {
				d.Set("port", rgp.NodeGroups[0].PrimaryEndpoint.Port)
			}

			if rgp.NodeGroups[0].PrimaryEndpoint.Address != nil {
				d.Set("primary_endpoint_address", rgp.NodeGroups[0].PrimaryEndpoint.Address)
			}
		}

		if rgp.NodeGroups[0].ReaderEndpoint != nil && rgp.NodeGroups[0].ReaderEndpoint.Address != nil {
			d.Set("reader_endpoint_address", rgp.NodeGroups[0].ReaderEndpoint.Address)
		}
	}

	d.Set("user_group_ids", rgp.UserGroupIds)

	// Tags cannot be read when the replication group is not Available
	log.Printf("[DEBUG] Waiting for ElastiCache Replication Group (%s) to become available", d.Id())

	_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("waiting for ElastiCache Replication Group to be available (%s): %w", aws.StringValue(rgp.ARN), err)
	}

	log.Printf("[DEBUG] Listing tags for ElastiCache Replication Group (%s)", d.Id())

	tags, err := ListTags(conn, aws.StringValue(rgp.ARN))

	if err != nil && !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		return fmt.Errorf("listing tags for ElastiCache Replication Group (%s): %w", aws.StringValue(rgp.ARN), err)
	}

	// tags not supported in all partitions
	if err != nil {
		log.Printf("[WARN] failed listing tags for ElastiCache Replication Group (%s): %s", aws.StringValue(rgp.ARN), err)
	}

	if tags != nil {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	}

	log.Printf("[DEBUG] ElastiCache Replication Group (%s): Checking underlying cache clusters", d.Id())

	// This section reads settings that require checking the underlying cache clusters
	if rgp.NodeGroups != nil && rgp.NodeGroups[0] != nil && len(rgp.NodeGroups[0].NodeGroupMembers) != 0 {
		cacheCluster := rgp.NodeGroups[0].NodeGroupMembers[0]

		res, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId:    cacheCluster.CacheClusterId,
			ShowCacheNodeInfo: aws.Bool(true),
		})
		if err != nil {
			return err
		}

		if len(res.CacheClusters) == 0 {
			return nil
		}

		c := res.CacheClusters[0]

		if err := setFromCacheCluster(d, c); err != nil {
			return err
		}

		d.Set("at_rest_encryption_enabled", c.AtRestEncryptionEnabled)
		d.Set("transit_encryption_enabled", c.TransitEncryptionEnabled)

		if c.AuthTokenEnabled != nil && !aws.BoolValue(c.AuthTokenEnabled) {
			d.Set("auth_token", nil)
		}
	}

	return nil
}

func resourceReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	if d.HasChanges(
		"cluster_mode.0.num_node_groups",
		"cluster_mode.0.replicas_per_node_group",
		"num_node_groups",
		"replicas_per_node_group",
	) {
		err := modifyReplicationGroupShardConfiguration(conn, d)
		if err != nil {
			return fmt.Errorf("error modifying ElastiCache Replication Group (%s) shard configuration: %w", d.Id(), err)
		}
	} else if d.HasChange("number_cache_clusters") {
		// TODO: remove when number_cache_clusters is removed from resource schema
		err := modifyReplicationGroupNumCacheClusters(conn, d, "number_cache_clusters")
		if err != nil {
			return fmt.Errorf("error modifying ElastiCache Replication Group (%s) clusters: %w", d.Id(), err)
		}
	} else if d.HasChange("num_cache_clusters") {
		err := modifyReplicationGroupNumCacheClusters(conn, d, "num_cache_clusters")
		if err != nil {
			return fmt.Errorf("error modifying ElastiCache Replication Group (%s) clusters: %w", d.Id(), err)
		}
	}

	requestUpdate := false
	params := &elasticache.ModifyReplicationGroupInput{
		ApplyImmediately:   aws.Bool(d.Get("apply_immediately").(bool)),
		ReplicationGroupId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		params.ReplicationGroupDescription = aws.String(d.Get("description").(string))
		requestUpdate = true
	}

	if d.HasChange("replication_group_description") {
		params.ReplicationGroupDescription = aws.String(d.Get("replication_group_description").(string))
		requestUpdate = true
	}

	if d.HasChange("automatic_failover_enabled") {
		params.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("auto_minor_version_upgrade") {
		v := d.Get("auto_minor_version_upgrade")
		if v, null, _ := nullable.Bool(v.(string)).Value(); !null {
			params.AutoMinorVersionUpgrade = aws.Bool(v)
		}
		requestUpdate = true
	}

	if d.HasChange("security_group_ids") {
		if attr := d.Get("security_group_ids").(*schema.Set); attr.Len() > 0 {
			params.SecurityGroupIds = flex.ExpandStringSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange("security_group_names") {
		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			params.CacheSecurityGroupNames = flex.ExpandStringSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange("log_delivery_configuration") {

		oldLogDeliveryConfig, newLogDeliveryConfig := d.GetChange("log_delivery_configuration")

		params.LogDeliveryConfigurations = []*elasticache.LogDeliveryConfigurationRequest{}
		logTypesToSubmit := make(map[string]bool)

		currentLogDeliveryConfig := newLogDeliveryConfig.(*schema.Set).List()
		for _, current := range currentLogDeliveryConfig {
			logDeliveryConfigurationRequest := expandLogDeliveryConfigurations(current.(map[string]interface{}))
			logTypesToSubmit[*logDeliveryConfigurationRequest.LogType] = true
			params.LogDeliveryConfigurations = append(params.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
		}

		previousLogDeliveryConfig := oldLogDeliveryConfig.(*schema.Set).List()
		for _, previous := range previousLogDeliveryConfig {
			logDeliveryConfigurationRequest := expandEmptyLogDeliveryConfigurations(previous.(map[string]interface{}))
			//if something was removed, send an empty request
			if !logTypesToSubmit[*logDeliveryConfigurationRequest.LogType] {
				params.LogDeliveryConfigurations = append(params.LogDeliveryConfigurations, &logDeliveryConfigurationRequest)
			}
		}
		requestUpdate = true
	}

	if d.HasChange("maintenance_window") {
		params.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("multi_az_enabled") {
		params.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("notification_topic_arn") {
		params.NotificationTopicArn = aws.String(d.Get("notification_topic_arn").(string))
		requestUpdate = true
	}

	if d.HasChange("parameter_group_name") {
		params.CacheParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("engine_version") {
		params.EngineVersion = aws.String(d.Get("engine_version").(string))
		requestUpdate = true
	}

	if d.HasChange("snapshot_retention_limit") {
		// This is a real hack to set the Snapshotting Cluster ID to be the first Cluster in the RG
		o, _ := d.GetChange("snapshot_retention_limit")
		if o.(int) == 0 {
			params.SnapshottingClusterId = aws.String(fmt.Sprintf("%s-001", d.Id()))
		}

		params.SnapshotRetentionLimit = aws.Int64(int64(d.Get("snapshot_retention_limit").(int)))
		requestUpdate = true
	}

	if d.HasChange("snapshot_window") {
		params.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		requestUpdate = true
	}

	if d.HasChange("node_type") {
		params.CacheNodeType = aws.String(d.Get("node_type").(string))
		requestUpdate = true
	}

	if d.HasChange("user_group_ids") {
		old, new := d.GetChange("user_group_ids")
		newSet := new.(*schema.Set)
		oldSet := old.(*schema.Set)
		add := newSet.Difference(oldSet)
		remove := oldSet.Difference(newSet)

		if add.Len() > 0 {
			params.UserGroupIdsToAdd = flex.ExpandStringSet(add)
			requestUpdate = true
		}

		if remove.Len() > 0 {
			params.UserGroupIdsToRemove = flex.ExpandStringSet(remove)
			requestUpdate = true
		}

	}

	if requestUpdate {
		_, err := conn.ModifyReplicationGroup(params)
		if err != nil {
			return fmt.Errorf("error updating ElastiCache Replication Group (%s): %w", d.Id(), err)
		}

		_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("auth_token") {
		params := &elasticache.ModifyReplicationGroupInput{
			ApplyImmediately:        aws.Bool(true),
			ReplicationGroupId:      aws.String(d.Id()),
			AuthTokenUpdateStrategy: aws.String("ROTATE"),
			AuthToken:               aws.String(d.Get("auth_token").(string)),
		}

		_, err := conn.ModifyReplicationGroup(params)
		if err != nil {
			return fmt.Errorf("error changing auth_token for ElastiCache Replication Group (%s): %w", d.Id(), err)
		}

		_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) auth_token change: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Get("arn").(string), o, n)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return fmt.Errorf("failed updating ElastiCache Replication Group (%s) tags: %w", d.Id(), err)
			}

			log.Printf("[WARN] failed updating tags for ElastiCache Replication Group (%s): %s", d.Id(), err)
		}
	}

	return resourceReplicationGroupRead(d, meta)
}

func resourceReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	v, hasGlobalReplicationGroupID := d.GetOk("global_replication_group_id")
	if hasGlobalReplicationGroupID {
		globalReplicationGroupID := v.(string)
		err := DisassociateReplicationGroup(conn, globalReplicationGroupID, d.Id(), meta.(*conns.AWSClient).Region, GlobalReplicationGroupDisassociationReadyTimeout)
		if err != nil {
			return fmt.Errorf("error disassociating ElastiCache Replication Group (%s) from Global Replication Group (%s): %w", d.Id(), globalReplicationGroupID, err)
		}
	}

	var finalSnapshotID = d.Get("final_snapshot_identifier").(string)
	err := deleteReplicationGroup(d.Id(), conn, finalSnapshotID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Replication Group (%s): %w", d.Id(), err)
	}

	if hasGlobalReplicationGroupID {
		paramGroupName := d.Get("parameter_group_name").(string)
		if paramGroupName != "" {
			err := deleteParameterGroup(conn, paramGroupName)
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("error deleting ElastiCache Parameter Group (%s): %w", d.Id(), err)
			}
		}
	}

	return nil
}

func DisassociateReplicationGroup(conn *elasticache.ElastiCache, globalReplicationGroupID, id, region string, readyTimeout time.Duration) error {
	input := &elasticache.DisassociateGlobalReplicationGroupInput{
		GlobalReplicationGroupId: aws.String(globalReplicationGroupID),
		ReplicationGroupId:       aws.String(id),
		ReplicationGroupRegion:   aws.String(region),
	}
	err := resource.Retry(readyTimeout, func() *resource.RetryError {
		_, err := conn.DisassociateGlobalReplicationGroup(input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DisassociateGlobalReplicationGroup(input)
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

	_, err = WaitGlobalReplicationGroupMemberDetached(conn, globalReplicationGroupID, id)
	if err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil

}

func deleteReplicationGroup(replicationGroupID string, conn *elasticache.ElastiCache, finalSnapshotID string, timeout time.Duration) error {
	input := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	if finalSnapshotID != "" {
		input.FinalSnapshotIdentifier = aws.String(finalSnapshotID)
	}

	// 10 minutes should give any creating/deleting cache clusters or snapshots time to complete
	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteReplicationGroup(input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
			return nil
		}
		// Cache Cluster is creating/deleting or Replication Group is snapshotting
		// InvalidReplicationGroupState: Cache cluster tf-acc-test-uqhe-003 is not in a valid state to be deleted
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidReplicationGroupStateFault) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteReplicationGroup(input)
	}

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = WaitReplicationGroupDeleted(conn, replicationGroupID, timeout)
	if err != nil {
		return err
	}

	return nil
}

func flattenNodeGroupsToClusterMode(nodeGroups []*elasticache.NodeGroup) []map[string]interface{} {
	if len(nodeGroups) == 0 {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"num_node_groups":         len(nodeGroups),
		"replicas_per_node_group": (len(nodeGroups[0].NodeGroupMembers) - 1),
	}
	return []map[string]interface{}{m}
}

func modifyReplicationGroupShardConfiguration(conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	if d.HasChange("cluster_mode.0.num_node_groups") {
		err := modifyReplicationGroupShardConfigurationNumNodeGroups(conn, d, "cluster_mode.0.num_node_groups")
		if err != nil {
			return err
		}
	}

	if d.HasChange("cluster_mode.0.replicas_per_node_group") {
		err := modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(conn, d, "cluster_mode.0.replicas_per_node_group")
		if err != nil {
			return err
		}
	}

	if d.HasChange("num_node_groups") {
		err := modifyReplicationGroupShardConfigurationNumNodeGroups(conn, d, "num_node_groups")
		if err != nil {
			return err
		}
	}

	if d.HasChange("replicas_per_node_group") {
		err := modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(conn, d, "replicas_per_node_group")
		if err != nil {
			return err
		}
	}

	return nil
}

func modifyReplicationGroupShardConfigurationNumNodeGroups(conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
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
	_, err := conn.ModifyReplicationGroupShardConfiguration(input)
	if err != nil {
		return fmt.Errorf("error modifying ElastiCache Replication Group shard configuration: %w", err)
	}

	_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) shard reconfiguration completion: %w", d.Id(), err)
	}

	return nil
}

func modifyReplicationGroupShardConfigurationReplicasPerNodeGroup(conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldReplicas := o.(int)
	newReplicas := n.(int)

	if newReplicas > oldReplicas {
		input := &elasticache.IncreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.IncreaseReplicaCount(input)
		if err != nil {
			return fmt.Errorf("error adding ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica addition: %w", d.Id(), err)
		}
	} else {
		input := &elasticache.DecreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.DecreaseReplicaCount(input)
		if err != nil {
			return fmt.Errorf("error removing ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = WaitReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica removal: %w", d.Id(), err)
		}
	}

	return nil
}

func modifyReplicationGroupNumCacheClusters(conn *elasticache.ElastiCache, d *schema.ResourceData, argument string) error {
	o, n := d.GetChange(argument)
	oldNumberCacheClusters := o.(int)
	newNumberCacheClusters := n.(int)

	var err error
	if newNumberCacheClusters > oldNumberCacheClusters {
		err = increaseReplicationGroupNumCacheClusters(conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	} else if newNumberCacheClusters < oldNumberCacheClusters {
		err = decreaseReplicationGroupNumCacheClusters(conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	}
	return err
}

func increaseReplicationGroupNumCacheClusters(conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.IncreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.IncreaseReplicaCount(input)
	if err != nil {
		return fmt.Errorf("error adding ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = WaitReplicationGroupMemberClustersAvailable(conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica addition: %w", replicationGroupID, err)
	}

	return nil
}

func decreaseReplicationGroupNumCacheClusters(conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.DecreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.DecreaseReplicaCount(input)
	if err != nil {
		return fmt.Errorf("error removing ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = WaitReplicationGroupMemberClustersAvailable(conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica removal: %w", replicationGroupID, err)
	}

	return nil
}

var validateReplicationGroupID schema.SchemaValidateFunc = validation.All(
	validation.StringLenBetween(1, 40),
	validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-]+$`), "must contain only alphanumeric characters and hyphens"),
	validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with a letter"),
	validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
	validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
)
