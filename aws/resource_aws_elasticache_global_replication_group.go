package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	elasticacheGlobalReplicationGroupRemovalTimeout = 2 * time.Minute
)

func resourceAwsElasticacheGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheGlobalReplicationGroupCreate,
		Read:   resourceAwsElasticacheGlobalReplicationGroupRead,
		Update: resourceAwsElasticacheGlobalReplicationGroupUpdate,
		Delete: resourceAwsElasticacheGlobalReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
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
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"global_replication_group_id_suffix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_replication_group_description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  false,
			},
			"global_replication_group_members": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"replication_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_group_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"primary_replication_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retain_primary_replication_group": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceAwsElasticacheGlobalReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	input := &elasticache.CreateGlobalReplicationGroupInput{
		GlobalReplicationGroupIdSuffix: aws.String(d.Get("global_replication_group_id_suffix").(string)),
		PrimaryReplicationGroupId:      aws.String(d.Get("primary_replication_group_id").(string)),
	}

	if v, ok := d.GetOk("global_replication_group_description"); ok {
		input.GlobalReplicationGroupDescription = aws.String(v.(string))
	}

	output, err := conn.CreateGlobalReplicationGroup(input)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Global Replication Group: %s", err)
	}

	d.SetId(aws.StringValue(output.GlobalReplicationGroup.GlobalReplicationGroupId))

	if err := waitForElasticacheGlobalReplicationGroupCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsElasticacheGlobalReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheGlobalReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	globalReplicationGroup, err := elasticacheDescribeGlobalReplicationGroup(conn, d.Id())

	if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ElastiCache Replication Group: %s", err)
	}

	if globalReplicationGroup == nil {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
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
	d.Set("engine_version", globalReplicationGroup.EngineVersion)
	d.Set("transit_encryption_enabled", globalReplicationGroup.TransitEncryptionEnabled)

	if err := d.Set("global_replication_group_members", flattenElasticacheGlobalReplicationGroupMembers(globalReplicationGroup.Members)); err != nil {
		return fmt.Errorf("error setting global_cluster_members: %w", err)
	}

	return nil
}

func resourceAwsElasticacheGlobalReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	input := &elasticache.ModifyGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(d.Get("apply_immediately").(bool)),
		GlobalReplicationGroupId: aws.String(d.Id()),
	}

	requestUpdate := false

	if d.HasChange("automatic_failover_enabled") {
		input.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("cache_node_type") {
		input.CacheNodeType = aws.String(d.Get("cache_node_type").(string))
		requestUpdate = true
	}

	if d.HasChange("engine_version") {
		input.EngineVersion = aws.String(d.Get("engine_version").(string))
		requestUpdate = true
	}

	if d.HasChange("global_replication_group_description") {
		input.GlobalReplicationGroupDescription = aws.String(d.Get("global_replication_group_description").(string))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.ModifyGlobalReplicationGroup(input)

		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error deleting ElastiCache Global Replication Group: %s", err)
		}

		if err := waitForElasticacheGlobalReplicationGroupUpdate(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ElastiCache Global Replcation Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsElasticacheGlobalReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	for _, globalReplicationGroupMemberRaw := range d.Get("global_replication_group_members").(*schema.Set).List() {
		globalReplicationGroupMember, ok := globalReplicationGroupMemberRaw.(map[string]interface{})

		if !ok {
			continue
		}
		replicationGroupID, ok := globalReplicationGroupMember["replication_group_id"].(string)
		if !ok {
			continue
		}

		role, ok := globalReplicationGroupMember["role"].(string)
		if !ok {
			continue
		}

		if role == "SECONDARY" {
			replicationGroupRegion, ok := globalReplicationGroupMember["replication_group_region"].(string)
			if !ok {
				continue
			}

			input := &elasticache.DisassociateGlobalReplicationGroupInput{
				GlobalReplicationGroupId: aws.String(d.Id()),
				ReplicationGroupId:       aws.String(replicationGroupID),
				ReplicationGroupRegion:   aws.String(replicationGroupRegion),
			}

			_, err := conn.DisassociateGlobalReplicationGroup(input)

			if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
				return nil
			}

			if err := waitForElasticacheGlobalReplicationGroupDisassociation(conn, d.Id(), replicationGroupID); err != nil {
				return fmt.Errorf("error waiting for Elasticache Replication Group (%s) removal from Elasticache Global Replication Group (%s): %w", replicationGroupID, d.Id(), err)
			}
		}
	}

	input := &elasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId:      aws.String(d.Id()),
		RetainPrimaryReplicationGroup: aws.Bool(d.Get("retain_primary_replication_group").(bool)),
	}

	log.Printf("[DEBUG] Deleting ElastiCache Global Replication Group (%s): %s", d.Id(), input)

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteGlobalReplicationGroup(input)

		if isAWSErr(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault, "is not empty") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteGlobalReplicationGroup(input)
	}

	if isAWSErr(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Global Replication Group: %s", err)
	}

	if err := waitForElasticacheGlobalReplicationGroupDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func flattenElasticacheGlobalReplicationGroupMembers(apiObjects []*elasticache.GlobalReplicationGroupMember) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"replication_group_id":     aws.StringValue(apiObject.ReplicationGroupId),
			"replication_group_region": aws.StringValue(apiObject.ReplicationGroupRegion),
			"role":                     aws.StringValue(apiObject.Role),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func elasticacheDescribeGlobalReplicationGroup(conn *elasticache.ElastiCache, globalReplicationGroupID string) (*elasticache.GlobalReplicationGroup, error) {
	var globalReplicationGroup *elasticache.GlobalReplicationGroup

	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(globalReplicationGroupID),
		ShowMemberInfo:           aws.Bool(true),
	}

	log.Printf("[DEBUG] Reading ElastiCache Global Replication Group (%s): %s", globalReplicationGroupID, input)
	err := conn.DescribeGlobalReplicationGroupsPages(input, func(page *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalReplicationGroups {
			if gc == nil {
				continue
			}

			if aws.StringValue(gc.GlobalReplicationGroupId) == globalReplicationGroupID {
				globalReplicationGroup = gc
				return false
			}
		}

		return !lastPage
	})

	return globalReplicationGroup, err
}

func elasticacheGlobalReplicationGroupRefreshFunc(conn *elasticache.ElastiCache, globalReplicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalReplicationGroup, err := elasticacheDescribeGlobalReplicationGroup(conn, globalReplicationGroupID)

		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			return nil, "deleted", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading ElastiCache Global Replication Group (%s): %s", globalReplicationGroupID, err)
		}

		if globalReplicationGroup == nil {
			return nil, "deleted", nil
		}

		return globalReplicationGroup, aws.StringValue(globalReplicationGroup.Status), nil
	}
}

func waitForElasticacheGlobalReplicationGroupCreation(conn *elasticache.ElastiCache, globalReplicationGroupID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"available", "primary-only"},
		Refresh: elasticacheGlobalReplicationGroupRefreshFunc(conn, globalReplicationGroupID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) availability", globalReplicationGroupID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForElasticacheGlobalReplicationGroupUpdate(conn *elasticache.ElastiCache, globalReplicationGroupID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"modifying"},
		Target:  []string{"available"},
		Refresh: elasticacheGlobalReplicationGroupRefreshFunc(conn, globalReplicationGroupID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) availability", globalReplicationGroupID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForElasticacheGlobalReplicationGroupDeletion(conn *elasticache.ElastiCache, globalReplicationGroupID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"primary-only",
			"modifying",
			"deleting",
		},
		Target:         []string{"deleted"},
		Refresh:        elasticacheGlobalReplicationGroupRefreshFunc(conn, globalReplicationGroupID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) deletion", globalReplicationGroupID)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}

func waitForElasticacheGlobalReplicationGroupDisassociation(conn *elasticache.ElastiCache, globalReplicationGroupID string, replicationGroupID string) error {
	stillExistsErr := fmt.Errorf("ElastiCache Replication Group still associated in ElastiCache Global Replication Group")
	var replicationGroup *elasticache.GlobalReplicationGroupMember

	err := resource.Retry(elasticacheGlobalReplicationGroupRemovalTimeout, func() *resource.RetryError {
		var err error

		replicationGroup, err = elasticacheDescribeGlobalReplicationGroupFromReplicationGroup(conn, globalReplicationGroupID, replicationGroupID)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if replicationGroup != nil {
			return resource.RetryableError(stillExistsErr)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = elasticacheDescribeGlobalReplicationGroupFromReplicationGroup(conn, globalReplicationGroupID, replicationGroupID)
	}

	if err != nil {
		return err
	}

	if replicationGroup != nil {
		return stillExistsErr
	}

	return nil
}

func elasticacheDescribeGlobalReplicationGroupFromReplicationGroup(conn *elasticache.ElastiCache, globalReplicationGroupID string, replicationGroupID string) (*elasticache.GlobalReplicationGroupMember, error) {
	globalReplicationGroup, err := elasticacheDescribeGlobalReplicationGroup(conn, globalReplicationGroupID)

	if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
		return nil, err
	}

	members := globalReplicationGroup.Members

	if len(members) == 0 {
		return nil, nil
	}

	for _, member := range members {
		if *member.ReplicationGroupId == replicationGroupID {
			return member, nil
		}
	}

	return nil, nil
}
