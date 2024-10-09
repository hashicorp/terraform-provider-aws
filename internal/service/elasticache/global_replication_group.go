// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	emptyDescription = " "
)

const (
	globalReplicationGroupRegionPrefixFormat = "[[:alpha:]]{5}-"
)

const (
	globalReplicationGroupMemberRolePrimary   = "PRIMARY"
	globalReplicationGroupMemberRoleSecondary = "SECONDARY"
)

// @SDKResource("aws_elasticache_global_replication_group", name="Global Replication Group")
func resourceGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalReplicationGroupCreate,
		ReadWithoutTimeout:   resourceGlobalReplicationGroupRead,
		UpdateWithoutTimeout: resourceGlobalReplicationGroupUpdate,
		DeleteWithoutTimeout: resourceGlobalReplicationGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				re := regexache.MustCompile("^" + globalReplicationGroupRegionPrefixFormat)
				d.Set("global_replication_group_id_suffix", re.ReplaceAllLiteralString(d.Id(), ""))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auth_token_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validRedisVersionString,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					if t, _ := regexp.MatchString(`[6-9]\.x`, new); t && old != "" {
						oldVersion, err := gversion.NewVersion(old)
						if err != nil {
							return false
						}
						return oldVersion.Segments()[0] >= 6
					}
					return false
				},
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_node_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"global_node_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slots": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_replication_group_id_suffix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_replication_group_description": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					if (old == emptyDescription && new == "") || (old == "" && new == emptyDescription) {
						return true
					}
					return false
				},
				StateFunc: func(v any) string {
					s := v.(string)
					if s == "" {
						return emptyDescription
					}
					return s
				},
			},
			// global_replication_group_members cannot be correctly implemented because any secondary
			// replication groups will be added after this resource completes.
			// "global_replication_group_members": {
			// 	Type:     schema.TypeSet,
			// 	Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"replication_group_id": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 			"replication_group_region": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 			"role": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 		},
			// 	},
			// },
			"num_node_groups": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"primary_replication_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateReplicationGroupID,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(globalReplicationGroupDefaultCreatedTimeout),
			Update: schema.DefaultTimeout(globalReplicationGroupDefaultUpdatedTimeout),
			Delete: schema.DefaultTimeout(globalReplicationGroupDefaultDeletedTimeout),
		},

		CustomizeDiff: customdiff.All(
			customizeDiffGlobalReplicationGroupEngineVersionErrorOnDowngrade,
			customizeDiffGlobalReplicationGroupParamGroupNameRequiresMajorVersionUpgrade,
			customdiff.ComputedIf("global_node_groups", diffHasChange("num_node_groups")),
		),
	}
}

func customizeDiffGlobalReplicationGroupEngineVersionErrorOnDowngrade(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	if diff.Id() == "" || !diff.HasChange(names.AttrEngineVersion) {
		return nil
	}

	if downgrade, err := engineVersionIsDowngrade(diff); err != nil {
		return err
	} else if !downgrade {
		return nil
	}

	return fmt.Errorf(`Downgrading ElastiCache Global Replication Group (%s) engine version requires replacement
of the Global Replication Group and all Replication Group members. The AWS provider cannot handle this internally.

Please use the "-replace" option on the terraform plan and apply commands (see https://www.terraform.io/cli/commands/plan#replace-address).`, diff.Id())
}

func customizeDiffGlobalReplicationGroupParamGroupNameRequiresMajorVersionUpgrade(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	return paramGroupNameRequiresMajorVersionUpgrade(diff)
}

// parameter_group_name can only be set when doing a major update,
// but we also should allow it to stay set afterwards
func paramGroupNameRequiresMajorVersionUpgrade(diff sdkv2.ResourceDiffer) error {
	o, n := diff.GetChange(names.AttrParameterGroupName)
	if o.(string) == n.(string) {
		return nil
	}

	if diff.Id() == "" {
		if !diff.HasChange(names.AttrEngineVersion) {
			return errors.New("cannot change parameter group name without upgrading major engine version")
		}
	}

	// cannot check for major version upgrade at plan time for new resource
	if diff.Id() != "" {
		o, n := diff.GetChange(names.AttrEngineVersion)

		newVersion, _ := normalizeEngineVersion(n.(string))
		oldVersion, _ := gversion.NewVersion(o.(string))

		vDiff := diffVersion(newVersion, oldVersion)
		if vDiff[0] == 0 && vDiff[1] == 0 {
			return errors.New("cannot change parameter group name without upgrading major engine version")
		}
		if vDiff[0] != 1 {
			return fmt.Errorf("cannot change parameter group name on minor engine version upgrade, upgrading from %s to %s", oldVersion.String(), newVersion.String())
		}
	}

	return nil
}

func resourceGlobalReplicationGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	id := d.Get("global_replication_group_id_suffix").(string)
	input := &elasticache.CreateGlobalReplicationGroupInput{
		GlobalReplicationGroupIdSuffix: aws.String(id),
		PrimaryReplicationGroupId:      aws.String(d.Get("primary_replication_group_id").(string)),
	}

	if v, ok := d.GetOk("global_replication_group_description"); ok {
		input.GlobalReplicationGroupDescription = aws.String(v.(string))
	}

	output, err := conn.CreateGlobalReplicationGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache Global Replication Group (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.GlobalReplicationGroup.GlobalReplicationGroupId))

	globalReplicationGroup, err := waitGlobalReplicationGroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Global Replication Group (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("automatic_failover_enabled"); ok {
		if v := v.(bool); v == flattenGlobalReplicationGroupAutomaticFailoverEnabled(globalReplicationGroup.Members) {
			log.Printf("[DEBUG] Not updating ElastiCache Global Replication Group (%s) automatic failover: no change from %t", d.Id(), v)
		} else {
			if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationAutomaticFailoverUpdater(v), "automatic failover", d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if v, ok := d.GetOk("cache_node_type"); ok {
		if v.(string) == aws.ToString(globalReplicationGroup.CacheNodeType) {
			log.Printf("[DEBUG] Not updating ElastiCache Global Replication Group (%s) node type: no change from %q", d.Id(), v)
		} else {
			if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupNodeTypeUpdater(v.(string)), "node type", d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		requestedVersion, _ := normalizeEngineVersion(v.(string))

		engineVersion, err := gversion.NewVersion(aws.ToString(globalReplicationGroup.EngineVersion))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache Global Replication Group (%s) engine version on creation: error reading engine version: %s", d.Id(), err)
		}

		diff := diffVersion(requestedVersion, engineVersion)

		if diff[0] == -1 || diff[1] == -1 { // Ignore patch version downgrade
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache Global Replication Group (%s) engine version on creation: cannot downgrade version when creating, is %s, want %s", d.Id(), engineVersion.String(), requestedVersion.String())
		}

		p := d.Get(names.AttrParameterGroupName).(string)

		if diff[0] == 1 {
			if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupEngineVersionMajorUpdater(v.(string), p), "engine version (major)", d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else if diff[1] == 1 {
			if p != "" {
				return sdkdiag.AppendErrorf(diags, "cannot change parameter group name on minor engine version upgrade, upgrading from %s to %s", engineVersion.String(), requestedVersion.String())
			}
			if t, _ := regexp.MatchString(`[6-9]\.x`, v.(string)); !t {
				if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupEngineVersionMinorUpdater(v.(string)), "engine version (minor)", d.Timeout(schema.TimeoutCreate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	if v, ok := d.GetOk("num_node_groups"); ok {
		if oldNodeGroupCount, newNodeGroupCount := len(globalReplicationGroup.GlobalNodeGroups), v.(int); newNodeGroupCount != oldNodeGroupCount {
			if newNodeGroupCount > oldNodeGroupCount {
				if err := increaseGlobalReplicationGroupNodeGroupCount(ctx, conn, d.Id(), newNodeGroupCount, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			} else if newNodeGroupCount < oldNodeGroupCount {
				ids := tfslices.ApplyToAll(globalReplicationGroup.GlobalNodeGroups, func(v awstypes.GlobalNodeGroup) string {
					return aws.ToString(v.GlobalNodeGroupId)
				})
				if err := decreaseGlobalReplicationGroupNodeGroupCount(ctx, conn, d.Id(), newNodeGroupCount, ids, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceGlobalReplicationGroupRead(ctx, d, meta)...)
}

func resourceGlobalReplicationGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	globalReplicationGroup, err := findGlobalReplicationGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): %s", d.Id(), err)
	}

	if status := aws.ToString(globalReplicationGroup.Status); !d.IsNewResource() && (status == globalReplicationGroupStatusDeleting || status == globalReplicationGroupStatusDeleted) {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) in deleted state (%s), removing from state", d.Id(), status)
		d.SetId("")
		return diags
	}

	d.Set(names.AttrARN, globalReplicationGroup.ARN)
	d.Set("at_rest_encryption_enabled", globalReplicationGroup.AtRestEncryptionEnabled)
	d.Set("auth_token_enabled", globalReplicationGroup.AuthTokenEnabled)
	d.Set("cache_node_type", globalReplicationGroup.CacheNodeType)
	d.Set("cluster_enabled", globalReplicationGroup.ClusterEnabled)
	d.Set(names.AttrEngine, globalReplicationGroup.Engine)
	d.Set("global_replication_group_description", globalReplicationGroup.GlobalReplicationGroupDescription)
	d.Set("global_replication_group_id", globalReplicationGroup.GlobalReplicationGroupId)
	d.Set("transit_encryption_enabled", globalReplicationGroup.TransitEncryptionEnabled)

	if err := setEngineVersionRedis(d, globalReplicationGroup.EngineVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Replication Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("global_node_groups", flattenGlobalNodeGroups(globalReplicationGroup.GlobalNodeGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_node_groups: %s", err)
	}
	d.Set("num_node_groups", len(globalReplicationGroup.GlobalNodeGroups))
	d.Set("automatic_failover_enabled", flattenGlobalReplicationGroupAutomaticFailoverEnabled(globalReplicationGroup.Members))

	d.Set("primary_replication_group_id", flattenGlobalReplicationGroupPrimaryGroupID(globalReplicationGroup.Members))

	return diags
}

func resourceGlobalReplicationGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	// Only one field can be changed per request
	if d.HasChange("cache_node_type") {
		if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupNodeTypeUpdater(d.Get("cache_node_type").(string)), "node type", d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("automatic_failover_enabled") {
		if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationAutomaticFailoverUpdater(d.Get("automatic_failover_enabled").(bool)), "automatic failover", d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrEngineVersion) {
		o, n := d.GetChange(names.AttrEngineVersion)

		newVersion, _ := normalizeEngineVersion(n.(string))
		oldVersion, _ := gversion.NewVersion(o.(string))

		diff := diffVersion(newVersion, oldVersion)
		if diff[0] == 1 {
			p := d.Get(names.AttrParameterGroupName).(string)
			if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupEngineVersionMajorUpdater(n.(string), p), "engine version (major)", d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else if diff[1] == 1 {
			if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupEngineVersionMinorUpdater(n.(string)), "engine version (minor)", d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("global_replication_group_description") {
		if err := updateGlobalReplicationGroup(ctx, conn, d.Id(), globalReplicationGroupDescriptionUpdater(d.Get("global_replication_group_description").(string)), names.AttrDescription, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("num_node_groups") {
		o, n := d.GetChange("num_node_groups")
		oldNodeGroupCount, newNodeGroupCount := o.(int), n.(int)

		if newNodeGroupCount != oldNodeGroupCount {
			if newNodeGroupCount > oldNodeGroupCount {
				if err := increaseGlobalReplicationGroupNodeGroupCount(ctx, conn, d.Id(), newNodeGroupCount, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			} else if newNodeGroupCount < oldNodeGroupCount {
				ids := tfslices.ApplyToAll(d.Get("global_node_groups").(*schema.Set).List(), func(tfMapRaw interface{}) string {
					tfMap := tfMapRaw.(map[string]interface{})
					return tfMap["global_node_group_id"].(string)
				})
				if err := decreaseGlobalReplicationGroupNodeGroupCount(ctx, conn, d.Id(), newNodeGroupCount, ids, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceGlobalReplicationGroupRead(ctx, d, meta)...)
}

func resourceGlobalReplicationGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	// Using Update timeout because the Global Replication Group could be in the middle of an update operation.
	if err := deleteGlobalReplicationGroup(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

type globalReplicationGroupUpdater func(input *elasticache.ModifyGlobalReplicationGroupInput)

func globalReplicationGroupDescriptionUpdater(description string) globalReplicationGroupUpdater {
	return func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.GlobalReplicationGroupDescription = aws.String(description)
	}
}

func globalReplicationGroupEngineVersionMinorUpdater(version string) globalReplicationGroupUpdater {
	return func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.EngineVersion = aws.String(version)
	}
}

func globalReplicationGroupEngineVersionMajorUpdater(version, paramGroupName string) globalReplicationGroupUpdater {
	return func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.EngineVersion = aws.String(version)
		input.CacheParameterGroupName = aws.String(paramGroupName)
	}
}

func globalReplicationAutomaticFailoverUpdater(enabled bool) globalReplicationGroupUpdater {
	return func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.AutomaticFailoverEnabled = aws.Bool(enabled)
	}
}

func globalReplicationGroupNodeTypeUpdater(nodeType string) globalReplicationGroupUpdater {
	return func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.CacheNodeType = aws.String(nodeType)
	}
}

func updateGlobalReplicationGroup(ctx context.Context, conn *elasticache.Client, id string, f globalReplicationGroupUpdater, propertyName string, timeout time.Duration) error {
	input := &elasticache.ModifyGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(true),
		GlobalReplicationGroupId: aws.String(id),
	}
	f(input)

	if _, err := conn.ModifyGlobalReplicationGroup(ctx, input); err != nil {
		return fmt.Errorf("updating ElastiCache Global Replication Group (%s) %s: %w", id, propertyName, err)
	}

	if _, err := waitGlobalReplicationGroupAvailable(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) update: %w", id, err)
	}

	return nil
}

func increaseGlobalReplicationGroupNodeGroupCount(ctx context.Context, conn *elasticache.Client, id string, newNodeGroupCount int, timeout time.Duration) error {
	input := &elasticache.IncreaseNodeGroupsInGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(true),
		GlobalReplicationGroupId: aws.String(id),
		NodeGroupCount:           aws.Int32(int32(newNodeGroupCount)),
	}

	_, err := conn.IncreaseNodeGroupsInGlobalReplicationGroup(ctx, input)

	if err != nil {
		return fmt.Errorf("increasing ElastiCache Global Replication Group (%s) node group count (%d): %w", id, newNodeGroupCount, err)
	}

	if _, err := waitGlobalReplicationGroupAvailable(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) update: %w", id, err)
	}

	return nil
}

func decreaseGlobalReplicationGroupNodeGroupCount(ctx context.Context, conn *elasticache.Client, id string, newNodeGroupCount int, nodeGroupIDs []string, timeout time.Duration) error {
	slices.SortFunc(nodeGroupIDs, func(a, b string) int {
		if globalReplicationGroupNodeNumber(a) < globalReplicationGroupNodeNumber(b) {
			return -1
		}
		if globalReplicationGroupNodeNumber(a) > globalReplicationGroupNodeNumber(b) {
			return 1
		}
		return 0
	})
	nodeGroupIDs = nodeGroupIDs[:newNodeGroupCount]

	input := &elasticache.DecreaseNodeGroupsInGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(true),
		GlobalNodeGroupsToRetain: nodeGroupIDs,
		GlobalReplicationGroupId: aws.String(id),
		NodeGroupCount:           aws.Int32(int32(newNodeGroupCount)),
	}

	_, err := conn.DecreaseNodeGroupsInGlobalReplicationGroup(ctx, input)

	if err != nil {
		return fmt.Errorf("decreasing ElastiCache Global Replication Group (%s) node group count (%d): %w", id, newNodeGroupCount, err)
	}

	if _, err := waitGlobalReplicationGroupAvailable(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) update: %w", id, err)
	}

	return nil
}

func deleteGlobalReplicationGroup(ctx context.Context, conn *elasticache.Client, id string, readyTimeout, deleteTimeout time.Duration) error {
	input := &elasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId:      aws.String(id),
		RetainPrimaryReplicationGroup: aws.Bool(true),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidGlobalReplicationGroupStateFault](ctx, readyTimeout, func() (interface{}, error) {
		return conn.DeleteGlobalReplicationGroup(ctx, input)
	})

	if errs.IsA[*awstypes.GlobalReplicationGroupNotFoundFault](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting ElastiCache Global Replication Group (%s): %w", id, err)
	}

	if _, err := waitGlobalReplicationGroupDeleted(ctx, conn, id, deleteTimeout); err != nil {
		return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) delete: %w", id, err)
	}

	return nil
}

func findGlobalReplicationGroupByID(ctx context.Context, conn *elasticache.Client, id string) (*awstypes.GlobalReplicationGroup, error) {
	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(id),
		ShowMemberInfo:           aws.Bool(true),
	}

	return findGlobalReplicationGroup(ctx, conn, input, tfslices.PredicateTrue[*awstypes.GlobalReplicationGroup]())
}

func findGlobalReplicationGroup(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeGlobalReplicationGroupsInput, filter tfslices.Predicate[*awstypes.GlobalReplicationGroup]) (*awstypes.GlobalReplicationGroup, error) {
	output, err := findGlobalReplicationGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGlobalReplicationGroups(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeGlobalReplicationGroupsInput, filter tfslices.Predicate[*awstypes.GlobalReplicationGroup]) ([]awstypes.GlobalReplicationGroup, error) {
	var output []awstypes.GlobalReplicationGroup

	pages := elasticache.NewDescribeGlobalReplicationGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.GlobalReplicationGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalReplicationGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusGlobalReplicationGroup(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGlobalReplicationGroupByID(ctx, conn, globalReplicationGroupID)

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
	globalReplicationGroupDefaultCreatedTimeout = 60 * time.Minute
	globalReplicationGroupDefaultUpdatedTimeout = 60 * time.Minute
	globalReplicationGroupDefaultDeletedTimeout = 20 * time.Minute
)

const (
	globalReplicationGroupStatusAvailable   = "available"
	globalReplicationGroupStatusCreating    = "creating"
	globalReplicationGroupStatusDeleted     = "deleted"
	globalReplicationGroupStatusDeleting    = "deleting"
	globalReplicationGroupStatusModifying   = "modifying"
	globalReplicationGroupStatusPrimaryOnly = "primary-only"
)

func waitGlobalReplicationGroupAvailable(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID string, timeout time.Duration) (*awstypes.GlobalReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{globalReplicationGroupStatusCreating, globalReplicationGroupStatusModifying},
		Target:     []string{globalReplicationGroupStatusAvailable, globalReplicationGroupStatusPrimaryOnly},
		Refresh:    statusGlobalReplicationGroup(ctx, conn, globalReplicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalReplicationGroup); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalReplicationGroupDeleted(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID string, timeout time.Duration) (*awstypes.GlobalReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			globalReplicationGroupStatusAvailable,
			globalReplicationGroupStatusPrimaryOnly,
			globalReplicationGroupStatusModifying,
			globalReplicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    statusGlobalReplicationGroup(ctx, conn, globalReplicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalReplicationGroup); ok {
		return output, err
	}

	return nil, err
}

func findGlobalReplicationGroupMemberByID(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID, replicationGroupID string) (*awstypes.GlobalReplicationGroupMember, error) {
	globalReplicationGroup, err := findGlobalReplicationGroupByID(ctx, conn, globalReplicationGroupID)

	if err != nil {
		return nil, err
	}

	if len(globalReplicationGroup.Members) == 0 {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	for _, v := range globalReplicationGroup.Members {
		if aws.ToString(v.ReplicationGroupId) == replicationGroupID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{
		Message: fmt.Sprintf("Replication Group (%s) not found in Global Replication Group (%s)", replicationGroupID, globalReplicationGroupID),
	}
}

func statusGlobalReplicationGroupMember(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID, replicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGlobalReplicationGroupMemberByID(ctx, conn, globalReplicationGroupID, replicationGroupID)

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
	globalReplicationGroupMemberStatusAssociated = "associated"
)

func waitGlobalReplicationGroupMemberDetached(ctx context.Context, conn *elasticache.Client, globalReplicationGroupID, replicationGroupID string, timeout time.Duration) (*awstypes.GlobalReplicationGroupMember, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{globalReplicationGroupMemberStatusAssociated},
		Target:     []string{},
		Refresh:    statusGlobalReplicationGroupMember(ctx, conn, globalReplicationGroupID, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalReplicationGroupMember); ok {
		return output, err
	}

	return nil, err
}

func flattenGlobalReplicationGroupAutomaticFailoverEnabled(members []awstypes.GlobalReplicationGroupMember) bool {
	if len(members) == 0 {
		return false
	}

	member := members[0]
	return member.AutomaticFailover == awstypes.AutomaticFailoverStatusEnabled
}

func flattenGlobalNodeGroups(nodeGroups []awstypes.GlobalNodeGroup) []any {
	if len(nodeGroups) == 0 {
		return nil
	}

	var l []any

	for _, nodeGroup := range nodeGroups {
		l = append(l, flattenGlobalNodeGroup(nodeGroup))
	}

	return l
}

func flattenGlobalNodeGroup(nodeGroup awstypes.GlobalNodeGroup) map[string]any {
	m := map[string]interface{}{}

	if v := nodeGroup.GlobalNodeGroupId; v != nil {
		m["global_node_group_id"] = aws.ToString(v)
	}

	if v := nodeGroup.Slots; v != nil {
		m["slots"] = aws.ToString(v)
	}

	return m
}

func flattenGlobalReplicationGroupPrimaryGroupID(members []awstypes.GlobalReplicationGroupMember) string {
	for _, member := range members {
		if aws.ToString(member.Role) == globalReplicationGroupMemberRolePrimary {
			return aws.ToString(member.ReplicationGroupId)
		}
	}
	return ""
}

func globalReplicationGroupNodeNumber(id string) int {
	re := regexache.MustCompile(`^.+-0{0,3}(\d+)$`)
	matches := re.FindStringSubmatch(id)
	if len(matches) == 2 {
		if v, err := strconv.Atoi(matches[1]); err == nil {
			return v
		}
	}
	return 0
}

func diffHasChange(key ...string) customdiff.ResourceConditionFunc {
	return func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
		return diff.HasChanges(key...)
	}
}
