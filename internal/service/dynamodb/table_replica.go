package dynamodb

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceTableReplica() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceTableReplicaCreate,
		Read:   resourceTableReplicaRead,
		Update: resourceTableReplicaUpdate,
		Delete: resourceTableReplicaDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(createTableTimeout),
			Delete: schema.DefaultTimeout(deleteTableTimeout),
			Update: schema.DefaultTimeout(updateTableTimeoutTotal),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": { // direct to replica
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_secondary_index": { // through global table
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"read_capacity_override": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"global_table_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"kms_key_arn": { // through global table
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"point_in_time_recovery": { // direct to replica
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"propagate_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"read_capacity_override": { // through global table
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"table_class_override": { // through global table
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dynamodb.TableClass_Values(), false),
			},
			"tags":     tftags.TagsSchema(),         // direct to replica
			"tags_all": tftags.TagsSchemaComputed(), // direct to replica
		},
	}
}

func resourceTableReplicaCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	globalRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Get("global_table_arn").(string), err)
	}

	if globalRegion == aws.StringValue(conn.Config.Region) {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Get("global_table_arn").(string), errors.New("replica cannot be in same region as global table"))
	}

	session, err := conns.NewSessionForRegion(&conn.Config, globalRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return fmt.Errorf("new session for region (%s): %w", globalRegion, err)
	}

	conn = dynamodb.New(session) // now global table region

	var replicaInput = &dynamodb.CreateReplicationGroupMemberAction{}

	replicaInput.RegionName = conn.Config.Region

	if v, ok := d.GetOk("kms_key_arn"); ok {
		replicaInput.KMSMasterKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("global_secondary_index"); ok && len(v.([]interface{})) > 0 {
		replicaInput.GlobalSecondaryIndexes = expandReplicaGlobalSecondaryIndexes(v.([]interface{}))
	}

	if v, ok := d.GetOk("read_capacity_override"); ok {
		replicaInput.ProvisionedThroughputOverride = &dynamodb.ProvisionedThroughputOverride{
			ReadCapacityUnits: aws.Int64(v.(int64)),
		}
	}

	if v, ok := d.GetOk("table_class_override"); ok {
		replicaInput.TableClassOverride = aws.String(v.(string))
	}

	tableName, err := TableNameFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return fmt.Errorf("creating replica of %s: %w", d.Get("global_table_arn").(string), err)
	}

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{
			{
				Create: replicaInput,
			},
		},
	}

	err = resource.Retry(maxDuration(replicaUpdateTimeout, d.Timeout(schema.TimeoutCreate)), func() *resource.RetryError {
		_, err := conn.UpdateTable(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTable(input)
	}

	if err != nil {
		return fmt.Errorf("creating replica (%s): %w", d.Get("global_table_arn").(string), err)
	}

	if _, err := waitReplicaActive(conn, tableName, meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for replica (%s) creation: %w", meta.(*conns.AWSClient).Region, err)
	}

	repARN, err := ARNForNewRegion(d.Get("global_table_arn").(string), aws.StringValue(conn.Config.Region))
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Get("global_table_arn").(string), err)
	}
	d.SetId(repARN)

	return resourceTableReplicaRead(d, meta)
}

func resourceTableReplicaRead(d *schema.ResourceData, meta interface{}) error {
	// handled through global table (main)
	// * global_secondary_index
	// * kms_key_arn
	// * read_capacity_override
	// * table_class_override
	conn := meta.(*conns.AWSClient).DynamoDBConn

	replicaRegion := aws.StringValue(conn.Config.Region)

	globalRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), err)
	}

	if globalRegion == replicaRegion {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), errors.New("replica cannot be in same region as global table"))
	}

	session, err := conns.NewSessionForRegion(&conn.Config, globalRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return fmt.Errorf("new session for region (%s): %w", globalRegion, err)
	}

	conn = dynamodb.New(session) // now global table region

	tableName, err := TableNameFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return fmt.Errorf("reading replica (%s): %w", d.Id(), err)
	}

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Dynamodb Table Replica (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), err)
	}

	if result == nil || result.Table == nil {
		if d.IsNewResource() {
			return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), errors.New("empty output after creation"))
		}
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id())
		d.SetId("")
		return nil
	}

	replica, err := filterReplicasByRegion(result.Table.Replicas, replicaRegion)

	if err := d.Set("global_secondary_index", flattenReplicaGlobalSecondaryIndexes(replica.GlobalSecondaryIndexes)); err != nil {
		return create.SettingError(names.DynamoDB, "Table Replica", d.Id(), "global_secondary_index", err)
	}

	d.Set("kms_key_arn", replica.KMSMasterKeyId)

	if replica.ProvisionedThroughputOverride != nil {
		d.Set("read_capacity_override", replica.ProvisionedThroughputOverride.ReadCapacityUnits)
	} else {
		d.Set("read_capacity_override", nil)
	}

	if replica.ReplicaTableClassSummary != nil {
		d.Set("table_class_override", replica.ReplicaTableClassSummary.TableClass)
	} else {
		d.Set("table_class_override", nil)
	}

	return resourceTableReplicaReadReplica(d, meta)
}

func resourceTableReplicaReadReplica(d *schema.ResourceData, meta interface{}) error {
	// handled direct to replica
	// * arn
	// * point_in_time_recovery
	// * tags
	conn := meta.(*conns.AWSClient).DynamoDBConn

	tableName, err := TableNameFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return fmt.Errorf("reading replica (%s): %w", d.Id(), err)
	}

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Dynamodb Table Replica (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), err)
	}

	if result == nil || result.Table == nil {
		if d.IsNewResource() {
			return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), errors.New("empty output after creation"))
		}
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", result.Table.TableArn)

	pitrOut, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil && !tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
		return create.Error(names.DynamoDB, create.ErrActionReading, "Table", d.Id(), fmt.Errorf("continuous backups: %w", err))
	}

	if pitrOut != nil && pitrOut.ContinuousBackupsDescription != nil && pitrOut.ContinuousBackupsDescription.ContinuousBackupsStatus != nil {
		d.Set("point_in_time_recovery", aws.StringValue(pitrOut.ContinuousBackupsDescription.ContinuousBackupsStatus) == dynamodb.PointInTimeRecoveryStatusEnabled)
	} else {
		d.Set("point_in_time_recovery", false)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil && !tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
		return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), fmt.Errorf("tags: %w", err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.SettingError(names.DynamoDB, "Table Replica", d.Id(), "tags", err)
	}

	if d.Get("propagate_tags").(bool) {
		globalRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
		if err != nil {
			return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), err)
		}

		session, err := conns.NewSessionForRegion(&conn.Config, globalRegion, meta.(*conns.AWSClient).TerraformVersion)
		if err != nil {
			return fmt.Errorf("new session for region (%s): %w", globalRegion, err)
		}

		conn = dynamodb.New(session) // now global table region

		globalTags, err := ListTags(conn, d.Get("global_table_arn").(string))

		if err != nil && !tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			return create.Error(names.DynamoDB, create.ErrActionReading, "Table Replica", d.Id(), fmt.Errorf("tags: %w", err))
		}

		globalTags = globalTags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
		globalTags = globalTags.RemoveDefaultConfig(defaultTagsConfig)

		tags = tags.Merge(globalTags)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.SettingError(names.DynamoDB, "Table Replica", d.Id(), "tags_all", err)
	}

	return nil
}

func resourceTableReplicaUpdate(d *schema.ResourceData, meta interface{}) error {
	// handled through global table (main)
	// * global_secondary_index
	// * kms_key_arn
	// * read_capacity_override
	// * table_class_override
	conn := meta.(*conns.AWSClient).DynamoDBConn

	tableName, err := TableNameFromARN(d.Id())
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionUpdating, "Table Replica", d.Id(), err)
	}

	replicaRegion := aws.StringValue(conn.Config.Region)

	globalRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), err)
	}

	if globalRegion == replicaRegion {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), errors.New("replica cannot be in same region as global table"))
	}

	session, err := conns.NewSessionForRegion(&conn.Config, globalRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return fmt.Errorf("new session for region (%s): %w", globalRegion, err)
	}

	conn = dynamodb.New(session) // now global table region

	viaGlobalChanges := false
	viaGlobalInput := &dynamodb.UpdateReplicationGroupMemberAction{
		RegionName: aws.String(replicaRegion),
	}

	if d.HasChange("global_secondary_index") {
		viaGlobalChanges = true
		viaGlobalInput.GlobalSecondaryIndexes = expandReplicaGlobalSecondaryIndexes(d.Get("global_secondary_index").(*schema.Set).List())
	}

	if d.HasChange("kms_key_arn") {
		viaGlobalChanges = true
		viaGlobalInput.KMSMasterKeyId = aws.String(d.Get("kms_key_arn").(string))
	}

	if d.HasChange("read_capacity_override") {
		viaGlobalChanges = true
		viaGlobalInput.ProvisionedThroughputOverride = &dynamodb.ProvisionedThroughputOverride{
			ReadCapacityUnits: aws.Int64(d.Get("read_capacity_override").(int64)),
		}
	}

	if d.HasChange("table_class_override") {
		viaGlobalChanges = true
		viaGlobalInput.TableClassOverride = aws.String(d.Get("table_class_override").(string))
	}

	if viaGlobalChanges {
		input := &dynamodb.UpdateTableInput{
			ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{
				{
					Update: viaGlobalInput,
				},
			},
			TableName: aws.String(tableName),
		}

		err := resource.Retry(maxDuration(replicaUpdateTimeout, d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			_, err := conn.UpdateTable(input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateTable(input)
		}

		if err != nil && !tfawserr.ErrMessageContains(err, "ValidationException", "no actions specified") {
			return fmt.Errorf("creating replica (%s): %w", d.Id(), err)
		}

		if _, err := waitReplicaActive(conn, tableName, replicaRegion, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for replica (%s) creation: %w", d.Id(), err)
		}
	}

	// handled direct to replica
	// * point_in_time_recovery
	// * tags
	conn = meta.(*conns.AWSClient).DynamoDBConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return create.Error(names.DynamoDB, create.ErrActionUpdating, "Table Replica", d.Id(), err)
		}
	}

	if d.HasChange("point_in_time_recovery") {
		if err := updatePITR(conn, tableName, d.Get("point_in_time_recovery").(bool), aws.StringValue(conn.Config.Region), meta.(*conns.AWSClient).TerraformVersion, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.Error(names.DynamoDB, create.ErrActionUpdating, "Table Replica", d.Id(), err)
		}
	}

	return resourceTableReplicaRead(d, meta)
}

func resourceTableReplicaDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	log.Printf("[DEBUG] DynamoDB delete Table Replica: %s", d.Id())

	tableName, err := TableNameFromARN(d.Id())
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionUpdating, "Table Replica", d.Id(), err)
	}

	replicaRegion := aws.StringValue(conn.Config.Region)

	globalRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.Error(names.DynamoDB, create.ErrActionCreating, "Table Replica", d.Id(), err)
	}

	session, err := conns.NewSessionForRegion(&conn.Config, globalRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return fmt.Errorf("new session for region (%s): %w", globalRegion, err)
	}

	conn = dynamodb.New(session) // now global table region

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{
			{
				Delete: &dynamodb.DeleteReplicationGroupMemberAction{
					RegionName: aws.String(replicaRegion),
				},
			},
		},
	}

	err = resource.Retry(updateTableTimeout, func() *resource.RetryError {
		_, err := conn.UpdateTable(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTable(input)
	}

	if err != nil {
		return fmt.Errorf("deleting replica (%s): %w", replicaRegion, err)
	}

	if _, err := waitReplicaDeleted(conn, tableName, replicaRegion, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for replica (%s) deletion: %w", replicaRegion, err)
	}

	return nil
}

func filterReplicasByRegion(replicas []*dynamodb.ReplicaDescription, region string) (*dynamodb.ReplicaDescription, error) {
	if len(replicas) == 0 {
		return nil, errors.New("no replicas found")
	}

	for _, replica := range replicas {
		if aws.StringValue(replica.RegionName) == region {
			return replica, nil
		}
	}

	return nil, errors.New("replica not found")
}

func expandReplicaGlobalSecondaryIndex(data map[string]interface{}) *dynamodb.ReplicaGlobalSecondaryIndex {
	if data == nil {
		return nil
	}

	idx := &dynamodb.ReplicaGlobalSecondaryIndex{
		IndexName: aws.String(data["name"].(string)),
	}

	if v, ok := data["read_capacity_override"].(int64); ok {
		idx.ProvisionedThroughputOverride = &dynamodb.ProvisionedThroughputOverride{
			ReadCapacityUnits: aws.Int64(v),
		}
	}

	return idx
}

func expandReplicaGlobalSecondaryIndexes(tfList []interface{}) []*dynamodb.ReplicaGlobalSecondaryIndex {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*dynamodb.ReplicaGlobalSecondaryIndex

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandReplicaGlobalSecondaryIndex(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenReplicaGlobalSecondaryIndex(apiObject *dynamodb.ReplicaGlobalSecondaryIndexDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.IndexName != nil {
		tfMap["name"] = aws.StringValue(apiObject.IndexName)
	}

	if apiObject.ProvisionedThroughputOverride != nil && apiObject.ProvisionedThroughputOverride.ReadCapacityUnits != nil {
		tfMap["read_capacity_override"] = aws.Int64Value(apiObject.ProvisionedThroughputOverride.ReadCapacityUnits)
	}

	return tfMap
}

func flattenReplicaGlobalSecondaryIndexes(apiObjects []*dynamodb.ReplicaGlobalSecondaryIndexDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenReplicaGlobalSecondaryIndex(apiObject))
	}

	return tfList
}
