package elasticache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	EmptyDescription = " "
)

const (
	GlobalReplicationGroupRegionPrefixFormat = "[[:alpha:]]{5}-"
)

const (
	GlobalReplicationGroupMemberRolePrimary   = "PRIMARY"
	GlobalReplicationGroupMemberRoleSecondary = "SECONDARY"
)

func ResourceGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlobalReplicationGroupCreate,
		Read:   resourceGlobalReplicationGroupRead,
		Update: resourceGlobalReplicationGroupUpdate,
		Delete: resourceGlobalReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				re := regexp.MustCompile("^" + GlobalReplicationGroupRegionPrefixFormat)
				d.Set("global_replication_group_id_suffix", re.ReplaceAllLiteralString(d.Id(), ""))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"cache_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
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
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: descriptionDiffSuppress,
				StateFunc:        descriptionStateFunc,
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
			"parameter_group_name": {
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

		CustomizeDiff: customdiff.Sequence(
			customizeDiffGlobalReplicationGroupEngineVersionErrorOnDowngrade,
			customizeDiffGlobalReplicationGroupParamGroupNameRequiresMajorVersionUpgrade,
		),
	}
}

func descriptionDiffSuppress(_, old, new string, _ *schema.ResourceData) bool {
	if (old == EmptyDescription && new == "") || (old == "" && new == EmptyDescription) {
		return true
	}
	return false
}

func descriptionStateFunc(v interface{}) string {
	s := v.(string)
	if s == "" {
		return EmptyDescription
	}
	return s
}

func customizeDiffGlobalReplicationGroupEngineVersionErrorOnDowngrade(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if diff.Id() == "" || !diff.HasChange("engine_version") {
		return nil
	}

	if downgrade, err := engineVersionIsDowngrade(diff); err != nil {
		return err
	} else if !downgrade {
		return nil
	}

	return fmt.Errorf(`Downgrading Elasticache Global Replication Group (%s) engine version requires replacement
of the Global Replication Group and all Replication Group members. The AWS provider cannot handle this internally.

Please use the "-replace" option on the terraform plan and apply commands (see https://www.terraform.io/cli/commands/plan#replace-address).`, diff.Id())
}

type changesDiffer interface {
	Id() string
	GetChange(key string) (interface{}, interface{})
	HasChange(key string) bool
}

func customizeDiffGlobalReplicationGroupParamGroupNameRequiresMajorVersionUpgrade(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	return paramGroupNameRequiresMajorVersionUpgrade(diff)
}

// parameter_group_name can only be set when doing a major update,
// but we also should allow it to stay set afterwards
func paramGroupNameRequiresMajorVersionUpgrade(diff changesDiffer) error {
	o, n := diff.GetChange("parameter_group_name")
	if o.(string) == n.(string) {
		return nil
	}

	if diff.Id() == "" {
		if !diff.HasChange("engine_version") {
			return errors.New("cannot change parameter group name without upgrading major engine version")
		}
	}

	// cannot check for major version upgrade at plan time for new resource
	if diff.Id() != "" {
		o, n := diff.GetChange("engine_version")

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

func resourceGlobalReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.CreateGlobalReplicationGroupInput{
		GlobalReplicationGroupIdSuffix: aws.String(d.Get("global_replication_group_id_suffix").(string)),
		PrimaryReplicationGroupId:      aws.String(d.Get("primary_replication_group_id").(string)),
	}

	if v, ok := d.GetOk("global_replication_group_description"); ok {
		input.GlobalReplicationGroupDescription = aws.String(v.(string))
	}

	output, err := conn.CreateGlobalReplicationGroup(input)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Global Replication Group: %w", err)
	}

	if output == nil || output.GlobalReplicationGroup == nil {
		return errors.New("error creating ElastiCache Global Replication Group: empty result")
	}

	d.SetId(aws.StringValue(output.GlobalReplicationGroup.GlobalReplicationGroupId))

	globalReplicationGroup, err := WaitGlobalReplicationGroupAvailable(conn, d.Id(), GlobalReplicationGroupDefaultCreatedTimeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) creation: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("engine_version"); ok {
		requestedVersion, _ := normalizeEngineVersion(v.(string))

		engineVersion, err := gversion.NewVersion(aws.StringValue(globalReplicationGroup.EngineVersion))
		if err != nil {
			return fmt.Errorf("error updating ElastiCache Global Replication Group (%s) engine version on creation: error reading engine version: %w", d.Id(), err)
		}

		diff := diffVersion(requestedVersion, engineVersion)

		if diff[0] == -1 || diff[1] == -1 { // Ignore patch version downgrade
			return fmt.Errorf("error updating ElastiCache Global Replication Group (%s) engine version on creation: cannot downgrade version when creating, is %s, want %s", d.Id(), engineVersion.String(), requestedVersion.String())
		}

		p := d.Get("parameter_group_name").(string)

		if diff[0] == 1 {
			err := updateGlobalReplicationGroup(conn, d.Id(), globalReplicationGroupEngineVersionMajorUpdater(v.(string), p))
			if err != nil {
				return fmt.Errorf("error updating ElastiCache Global Replication Group (%s) engine version on creation: %w", d.Id(), err)
			}
		} else if diff[1] == 1 {
			if p != "" {
				return fmt.Errorf("cannot change parameter group name on minor engine version upgrade, upgrading from %s to %s", engineVersion.String(), requestedVersion.String())
			}
			if t, _ := regexp.MatchString(`[6-9]\.x`, v.(string)); !t {
				err := updateGlobalReplicationGroup(conn, d.Id(), globalReplicationGroupEngineVersionMinorUpdater(v.(string)))
				if err != nil {
					return fmt.Errorf("error updating ElastiCache Global Replication Group (%s) engine version on creation: %w", d.Id(), err)
				}
			}
		}
	}

	return resourceGlobalReplicationGroupRead(d, meta)
}

func resourceGlobalReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	globalReplicationGroup, err := FindGlobalReplicationGroupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Replication Group: %w", err)
	}

	if aws.StringValue(globalReplicationGroup.Status) == "deleting" || aws.StringValue(globalReplicationGroup.Status) == "deleted" {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(globalReplicationGroup.Status))
		d.SetId("")
		return nil
	}

	d.Set("arn", globalReplicationGroup.ARN)
	d.Set("at_rest_encryption_enabled", globalReplicationGroup.AtRestEncryptionEnabled)
	d.Set("auth_token_enabled", globalReplicationGroup.AuthTokenEnabled)
	d.Set("cache_node_type", globalReplicationGroup.CacheNodeType)
	d.Set("cluster_enabled", globalReplicationGroup.ClusterEnabled)
	d.Set("engine", globalReplicationGroup.Engine)
	d.Set("global_replication_group_description", globalReplicationGroup.GlobalReplicationGroupDescription)
	d.Set("global_replication_group_id", globalReplicationGroup.GlobalReplicationGroupId)
	d.Set("transit_encryption_enabled", globalReplicationGroup.TransitEncryptionEnabled)

	err = setEngineVersionRedis(d, globalReplicationGroup.EngineVersion)
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Replication Group: %w", err)
	}

	d.Set("primary_replication_group_id", flattenGlobalReplicationGroupPrimaryGroupID(globalReplicationGroup.Members))

	return nil
}

type globalReplicationGroupUpdater func(input *elasticache.ModifyGlobalReplicationGroupInput)

func resourceGlobalReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	if d.HasChange("engine_version") {
		o, n := d.GetChange("engine_version")

		newVersion, _ := normalizeEngineVersion(n.(string))
		oldVersion, _ := gversion.NewVersion(o.(string))

		diff := diffVersion(newVersion, oldVersion)
		if diff[0] == 1 {
			p := d.Get("parameter_group_name").(string)
			err := updateGlobalReplicationGroup(conn, d.Id(), globalReplicationGroupEngineVersionMajorUpdater(n.(string), p))
			if err != nil {
				return fmt.Errorf("error updating ElastiCache Global Replication Group (%s): %w", d.Id(), err)
			}
		} else if diff[1] == 1 {
			err := updateGlobalReplicationGroup(conn, d.Id(), globalReplicationGroupEngineVersionMinorUpdater(n.(string)))
			if err != nil {
				return fmt.Errorf("error updating ElastiCache Global Replication Group (%s): %w", d.Id(), err)
			}
		}
	}

	// Only one field can be changed per request
	updaters := map[string]globalReplicationGroupUpdater{}
	updaters["global_replication_group_description"] = func(input *elasticache.ModifyGlobalReplicationGroupInput) {
		input.GlobalReplicationGroupDescription = aws.String(d.Get("global_replication_group_description").(string))
	}

	for k, f := range updaters {
		if d.HasChange(k) {
			if err := updateGlobalReplicationGroup(conn, d.Id(), f); err != nil {
				return fmt.Errorf("error updating ElastiCache Global Replication Group (%s): %w", d.Id(), err)
			}
		}
	}

	return resourceGlobalReplicationGroupRead(d, meta)
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

func updateGlobalReplicationGroup(conn *elasticache.ElastiCache, id string, f globalReplicationGroupUpdater) error {
	input := &elasticache.ModifyGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(true),
		GlobalReplicationGroupId: aws.String(id),
	}
	f(input)

	if _, err := conn.ModifyGlobalReplicationGroup(input); err != nil {
		return err
	}

	if _, err := WaitGlobalReplicationGroupAvailable(conn, id, GlobalReplicationGroupDefaultUpdatedTimeout); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func resourceGlobalReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	// Using Update timeout because the Global Replication Group could be in the middle of an update operation
	err := DeleteGlobalReplicationGroup(conn, d.Id(), GlobalReplicationGroupDefaultUpdatedTimeout)
	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Global Replication Group: %w", err)
	}

	return nil
}

func DeleteGlobalReplicationGroup(conn *elasticache.ElastiCache, id string, readyTimeout time.Duration) error {
	input := &elasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId:      aws.String(id),
		RetainPrimaryReplicationGroup: aws.Bool(true),
	}

	err := resource.Retry(readyTimeout, func() *resource.RetryError {
		_, err := conn.DeleteGlobalReplicationGroup(input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
			return resource.NonRetryableError(&resource.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
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
		_, err = conn.DeleteGlobalReplicationGroup(input)
	}
	if tfresource.NotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if _, err := WaitGlobalReplicationGroupDeleted(conn, id); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func flattenGlobalReplicationGroupPrimaryGroupID(members []*elasticache.GlobalReplicationGroupMember) string {
	for _, member := range members {
		if aws.StringValue(member.Role) == GlobalReplicationGroupMemberRolePrimary {
			return aws.StringValue(member.ReplicationGroupId)
		}
	}
	return ""
}
