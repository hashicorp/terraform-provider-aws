package elasticache

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
			// Leaving space for `engine_version` for creation and updating.
			// `engine_version` cannot be used for returning the version because, starting with Redis 6,
			// version configuration is major-version-only: `engine_version = "6.x"` or major-minor-version-only: `engine_version = "6.2"`,
			// while `engine_version_actual` will be the full version e.g. `6.0.5`
			// See also https://github.com/hashicorp/terraform-provider-aws/issues/15625
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
	}
}

func descriptionDiffSuppress(_, old, new string, d *schema.ResourceData) bool {
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

	d.SetId(aws.StringValue(output.GlobalReplicationGroup.GlobalReplicationGroupId))

	if _, err := WaitGlobalReplicationGroupAvailable(conn, d.Id(), GlobalReplicationGroupDefaultCreatedTimeout); err != nil {
		return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) creation: %w", d.Id(), err)
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
	d.Set("engine_version_actual", globalReplicationGroup.EngineVersion)
	d.Set("global_replication_group_description", globalReplicationGroup.GlobalReplicationGroupDescription)
	d.Set("global_replication_group_id", globalReplicationGroup.GlobalReplicationGroupId)
	d.Set("transit_encryption_enabled", globalReplicationGroup.TransitEncryptionEnabled)

	d.Set("primary_replication_group_id", flattenGlobalReplicationGroupPrimaryGroupID(globalReplicationGroup.Members))

	return nil
}

func resourceGlobalReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	// Only one field can be changed per request
	updaters := map[string]globalReplicationGroupUpdater{}
	if !d.IsNewResource() {
		updaters["global_replication_group_description"] = func(input *elasticache.ModifyGlobalReplicationGroupInput) {
			input.GlobalReplicationGroupDescription = aws.String(d.Get("global_replication_group_description").(string))
		}
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

type globalReplicationGroupUpdater func(input *elasticache.ModifyGlobalReplicationGroupInput)

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
