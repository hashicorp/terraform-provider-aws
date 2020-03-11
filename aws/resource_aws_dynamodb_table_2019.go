package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func replicaSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"region_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"kms_master_key_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"provision_capacity_override": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"read_capacity": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
				"global_secondary_index": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"provisioned_capacity_override": {
								Type:     schema.TypeMap,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"read_capacity": {
											Type:     schema.TypeInt,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsDynamoDbTable2019() *schema.Resource {
	schema := resourceAwsDynamoDbTable()
	schema.Create = resourceAwsDynamoDbTable2019Create
	schema.Read = resourceAwsDynamoDbTable2019Read
	schema.Update = resourceAwsDynamoDbTable2019Update
	schema.Delete = resourceAwsDynamoDbTable2019Delete
	schema.Schema["replica"] = replicaSchema()

	return schema
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
