package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsDynamoDbTable2019() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTable2019Create,
		Read:   resourceAwsDynamoDbTable2019Read,
		Update: resourceAwsDynamoDbTable2019Update,
		Delete: resourceAwsDynamoDbTable2019Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			func(diff *schema.ResourceDiff, v interface{}) error {
				return validateDynamoDbStreamSpec(diff)
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				return validateDynamoDbTableAttributes(diff)
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				if diff.Id() != "" && diff.HasChange("server_side_encryption") {
					o, n := diff.GetChange("server_side_encryption")
					if isDynamoDbTableOptionDisabled(o) && isDynamoDbTableOptionDisabled(n) {
						return diff.Clear("server_side_encryption")
					}
				}
				return nil
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				if diff.Id() != "" && diff.HasChange("point_in_time_recovery") {
					o, n := diff.GetChange("point_in_time_recovery")
					if isDynamoDbTableOptionDisabled(o) && isDynamoDbTableOptionDisabled(n) {
						return diff.Clear("point_in_time_recovery")
					}
				}
				return nil
			},
		),

		SchemaVersion: 1,
		MigrateState:  resourceAwsDynamoDbTableMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hash_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"range_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"billing_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  dynamodb.BillingModeProvisioned,
				ValidateFunc: validation.StringInSlice([]string{
					dynamodb.BillingModePayPerRequest,
					dynamodb.BillingModeProvisioned,
				}, false),
			},
			"write_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"attribute": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								dynamodb.ScalarAttributeTypeB,
								dynamodb.ScalarAttributeTypeN,
								dynamodb.ScalarAttributeTypeS,
							}, false),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"ttl": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"local_secondary_index": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"global_secondary_index": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"write_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"read_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"hash_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"stream_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"stream_view_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToUpper(value)
				},
				ValidateFunc: validation.StringInSlice([]string{
					"",
					dynamodb.StreamViewTypeNewImage,
					dynamodb.StreamViewTypeOldImage,
					dynamodb.StreamViewTypeNewAndOldImages,
					dynamodb.StreamViewTypeKeysOnly,
				}, false),
			},
			"stream_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_side_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"point_in_time_recovery": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"replica": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsDynamoDbTable2019Create(d *schema.ResourceData, meta interface{}) error {
	err := resourceAwsDynamoDbTableCreate(d, meta)
	if err != nil {
		return fmt.Errorf("error creating DynamoDB Table (%s) %s", d.Id(), err)
	}
	conn := meta.(*AWSClient).dynamodbconn

	if _, ok := d.GetOk("replica"); ok {
		if err := createDynamoDbReplicas(d.Id(), d.Get("replica").([]interface{}), conn); err != nil {
			return fmt.Errorf("error enabled DynamoDB Table (%s) replicas: %s", d.Id(), err)
		}
	}

	if err := waitForDynamoDbTableToBeActive(d.Id(), d.Timeout(schema.TimeoutCreate), conn); err != nil {
		return err
	}

	return resourceAwsDynamoDbTable2019Read(d, meta)
}

func createDynamoDbReplicas(tableName string, replicas []interface{}, conn *dynamodb.DynamoDB) error {
	for _, replica := range replicas {
		var ops []*dynamodb.ReplicationGroupUpdate
		if regionName, ok := replica.(map[string]interface{})["region_name"]; ok {
			ops = append(ops, &dynamodb.ReplicationGroupUpdate{
				Create: &dynamodb.CreateReplicationGroupMemberAction{
					RegionName: aws.String(regionName.(string)),
				},
			})

			input := &dynamodb.UpdateTableInput{
				TableName:      aws.String(tableName),
				ReplicaUpdates: ops,
			}

			log.Printf("[DEBUG] Updating DynamoDB Replicas to %v", input)

			err := resource.Retry(20*time.Minute, func() *resource.RetryError {
				_, err := conn.UpdateTable(input)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.UpdateTable(input)
			}
			if err != nil {
				return fmt.Errorf("Error updating DynamoDB Replicas status: %s", err)
			}

			if err := waitForDynamoDbReplicaUpdateToBeCompleted(tableName, regionName.(string), 20*time.Minute, conn); err != nil {
				return fmt.Errorf("Error waiting for DynamoDB replica update: %s", err)
			}
		}
	}
	return nil
}

func deleteDynamoDbReplicas(tableName string, replicas []interface{}, conn *dynamodb.DynamoDB) error {
	for _, replica := range replicas {
		var ops []*dynamodb.ReplicationGroupUpdate
		if regionName, ok := replica.(map[string]interface{})["region_name"]; ok {
			ops = append(ops, &dynamodb.ReplicationGroupUpdate{
				Delete: &dynamodb.DeleteReplicationGroupMemberAction{
					RegionName: aws.String(regionName.(string)),
				},
			})

			input := &dynamodb.UpdateTableInput{
				TableName:      aws.String(tableName),
				ReplicaUpdates: ops,
			}

			log.Printf("[DEBUG] Deleting DynamoDB Replicas to %v", input)

			err := resource.Retry(20*time.Minute, func() *resource.RetryError {
				_, err := conn.UpdateTable(input)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.UpdateTable(input)
			}
			if err != nil {
				return fmt.Errorf("Error deleting DynamoDB Replicas status: %s", err)
			}

			if err := waitForDynamoDbReplicaDeleteToBeCompleted(tableName, regionName.(string), 20*time.Minute, conn); err != nil {
				return fmt.Errorf("Error waiting for DynamoDB replica delete: %s", err)
			}
		}
	}
	return nil
}

func resourceAwsDynamoDbTable2019Update(d *schema.ResourceData, meta interface{}) error {
	err := resourceAwsDynamoDbTableUpdate(d, meta)
	if err != nil {
		return fmt.Errorf("error updating DynamoDB Table (%s) %s", d.Id(), err)
	}

	conn := meta.(*AWSClient).dynamodbconn

	if d.HasChange("replica") {
		var replicaUpdates []*dynamodb.ReplicationGroupUpdate
		o, n := d.GetChange("replica")

		replicaUpdates, _ = diffDynamoDbReplicas(o.([]interface{}), n.([]interface{}))
		log.Printf("[DEBUG] replica updates %s", replicaUpdates)
		for _, replicaUpdate := range replicaUpdates {
			var ops []*dynamodb.ReplicationGroupUpdate
			ops = append(ops, replicaUpdate)

			replicaInput := &dynamodb.UpdateTableInput{
				TableName:      aws.String(d.Id()),
				ReplicaUpdates: ops,
			}
			replicaInput.ReplicaUpdates = replicaUpdates
			_, replicaErr := conn.UpdateTable(replicaInput)
			if replicaErr == nil {
				if replicaUpdate.Delete == nil {
					log.Printf("[DEBUG] waiting for replica to be updated")
					waitForDynamoDbReplicaUpdateToBeCompleted(d.Id(), aws.StringValue(replicaUpdate.Update.RegionName), 20*time.Minute, conn)
				} else {
					log.Printf("[DEBUG] waiting for replica to be deleted")
					waitForDynamoDbReplicaDeleteToBeCompleted(d.Id(), aws.StringValue(replicaUpdate.Delete.RegionName), 20*time.Minute, conn)
				}
			} else {
				return fmt.Errorf("error updating DynamoDB Table (%s): %s", d.Id(), replicaErr)
			}
		}
	}

	return resourceAwsDynamoDbTable2019Read(d, meta)
}

func resourceAwsDynamoDbTable2019Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	err = flattenAwsDynamoDbTableResource_2019(d, result.Table)
	if err != nil {
		return err
	}

	ttlOut, err := conn.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing DynamoDB Table (%s) Time to Live: %s", d.Id(), err)
	}
	if err := d.Set("ttl", flattenDynamoDbTtl(ttlOut)); err != nil {
		return fmt.Errorf("error setting ttl: %s", err)
	}

	tags, err := readDynamoDbTableTags(d.Get("arn").(string), conn)
	if err != nil {
		return err
	}

	d.Set("tags", tags)

	pitrOut, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	})
	if err != nil && !isAWSErr(err, "UnknownOperationException", "") {
		return err
	}
	d.Set("point_in_time_recovery", flattenDynamoDbPitr(pitrOut))

	return nil
}

func resourceAwsDynamoDbTable2019Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	log.Printf("[DEBUG] DynamoDB delete table: %s", d.Id())

	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(d.Id()),
	}

	output, err := conn.DescribeTable(input)
	log.Printf("[DEBUG] DynamoDB delete describe: %s", output)

	if len(output.Table.Replicas) > 0 {
		if err := deleteDynamoDbReplicas(d.Id(), d.Get("replica").([]interface{}), conn); err != nil {
			return fmt.Errorf("error enabled DynamoDB Table (%s) replicas: %s", d.Id(), err)
		}
	}

	err = deleteAwsDynamoDbTable(d.Id(), conn)
	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "Requested resource not found: Table: ") {
			return nil
		}
		return fmt.Errorf("error deleting DynamoDB Table (%s): %s", d.Id(), err)
	}

	if err := waitForDynamodbTableDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for DynamoDB Table (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func waitForDynamoDbReplicaUpdateToBeCompleted(tableName string, region string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
		},
		Target: []string{
			dynamodb.ReplicaStatusActive,
		},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}
			log.Printf("[DEBUG] DynamoDB replicas: %s", result.Table.Replicas)

			if len(result.Table.Replicas) == 0 {
				return result, dynamodb.ReplicaStatusCreating, nil
			}
			// Find replica
			var targetReplica *dynamodb.ReplicaDescription
			for _, replica := range result.Table.Replicas {
				if *replica.RegionName == region {
					targetReplica = replica
				}
			}

			if targetReplica == nil {
				return nil, "", nil
			}

			return result, aws.StringValue(targetReplica.ReplicaStatus), nil
		},
	}
	_, err := stateConf.WaitForState()

	return err
}

func waitForDynamoDbReplicaDeleteToBeCompleted(tableName string, region string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			dynamodb.ReplicaStatusActive,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			log.Printf("[DEBUG] all replicas for waiting: %s", result.Table.Replicas)
			if len(result.Table.Replicas) == 0 {
				return result, "", nil
			}

			// Find replica
			var targetReplica *dynamodb.ReplicaDescription
			for _, replica := range result.Table.Replicas {
				if *replica.RegionName == region {
					targetReplica = replica
				}
			}
			log.Printf("[DEBUG] targetReplica: %s", targetReplica)

			if targetReplica == nil {
				return result, "", nil
			}

			return result, aws.StringValue(targetReplica.ReplicaStatus), nil
		},
	}
	_, err := stateConf.WaitForState()

	return err
}
