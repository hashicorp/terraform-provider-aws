package aws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsDynamoDbTableGsi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableGsiCreate,
		Read:   resourceAwsDynamoDbTableGsiRead,
		Update: resourceAwsDynamoDbTableGsiUpdate,
		Delete: resourceAwsDynamoDbTableGsiDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		// ???
		// CustomizeDiff: customdiff.Sequence(
		// 	func(diff *schema.ResourceDiff, v interface{}) error {
		// 		return validateDynamoDbTableAttributes(diff)
		// 	},
		// ),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"table_name": {
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
				Required: true,
				ForceNew: true,
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
			"write_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
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
	}
}

func resourceAwsDynamoDbTableGsiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	keySchemaMap := map[string]interface{}{
		"hash_key": d.Get("hash_key").(string),
	}
	if v, ok := d.GetOk("range_key"); ok {
		keySchemaMap["range_key"] = v.(string)
	}

	log.Printf("[DEBUG] Creating DynamoDB table index with key schema: %#v", keySchemaMap)
	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
	}

	// TODO: replace with keySchema
	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	projection := &dynamodb.Projection{
		ProjectionType: aws.String(d.Get("projection_type").(string)),
	}

	if v, ok := d.GetOk("non_key_attributes"); ok && len(v.([]interface{})) > 0 {
		projection.NonKeyAttributes = expandStringList(v.([]interface{}))
	}

	createOp := &dynamodb.GlobalSecondaryIndexUpdate{
		Create: &dynamodb.CreateGlobalSecondaryIndexAction{
			IndexName: aws.String(d.Id()),
			KeySchema: expandDynamoDbKeySchema(keySchemaMap),
			ProvisionedThroughput: expandDynamoDbProvisionedThroughput(map[string]interface{}{
				"read_capacity":  d.Get("read_capacity"),
				"write_capacity": d.Get("write_capacity"),
			}),
			Projection: projection,
		},
	}
	req.GlobalSecondaryIndexUpdates = []*dynamodb.GlobalSecondaryIndexUpdate{createOp}

	var output *dynamodb.UpdateTableOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.UpdateTable(req)
		if err != nil {
			if isAWSErr(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			// TODO: double check error codes for GSI creation
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "indexed tables that can be created simultaneously") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// TODO: what ouputs do we get from GSI creation?
	// d.SetId(*output.TableDescription.TableName)

	if err := waitForDynamoDbGSIToBeActive(d.Get("table_name").(string), d.Id(), conn); err != nil {
		return err
	}

	// ???
	return resourceAwsDynamoDbTableGsiUpdate(d, meta)
}

func resourceAwsDynamoDbTableGsiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
					IndexName: aws.String(d.Id()),
					ProvisionedThroughput: expandDynamoDbProvisionedThroughput(map[string]interface{}{
						"read_capacity":  d.Get("read_capacity"),
						"write_capacity": d.Get("write_capacity"),
					}),
				},
			},
		},
	}

	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	log.Printf("[DEBUG] Updating DynamoDB table GSI: %s", req)
	var output *dynamodb.UpdateTableOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.UpdateTable(req)
		if err != nil {
			if isAWSErr(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			// TODO: double check error codes for GSI creation
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "indexed tables that can be created simultaneously") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := waitForDynamoDbGSIToBeActive(d.Get("table_name").(string), d.Id(), conn); err != nil {
		return err
	}

	return resourceAwsDynamoDbTableGsiRead(d, meta)
}

func resourceAwsDynamoDbTableGsiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn
	tableName := d.Get("table_name").(string)

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		return err
	}

	err = flattenAwsDynamoDbTableGsiResource(d, result.Table)
	if err != nil {
		return err
	}

	return nil
}

func resourceAwsDynamoDbTableGsiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	log.Printf("[DEBUG] DynamoDB delete index: %s", d.Id())

	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
					IndexName: aws.String(d.Id()),
				},
			},
		},
	}

	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.UpdateTable(req)
		if err != nil {
			// Subscriber limit exceeded: Only 10 tables can be created, updated, or deleted simultaneously
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "Requested resource not found: ") {
				return resource.NonRetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	err := waitForDynamoDbGSIToBeActive(d.Get("table_name").(string), d.Id(), conn)

	return err
}

// Helpers

func flattenAwsDynamoDbTableGsiResource(d *schema.ResourceData, table *dynamodb.TableDescription) error {
	attributes := []interface{}{}
	for _, attrdef := range table.AttributeDefinitions {
		attribute := map[string]string{
			"name": *attrdef.AttributeName,
			"type": *attrdef.AttributeType,
		}
		attributes = append(attributes, attribute)
	}

	d.Set("attribute", attributes)
	d.Set("table_name", table.TableName)

	for _, gsiObject := range table.GlobalSecondaryIndexes {
		if *gsiObject.IndexName == d.Id() {
			d.Set("write_capacity", gsiObject.ProvisionedThroughput.WriteCapacityUnits)
			d.Set("read_capacity", gsiObject.ProvisionedThroughput.ReadCapacityUnits)
			d.Set("projection_type", gsiObject.Projection.ProjectionType)

			for _, attribute := range gsiObject.KeySchema {
				if *attribute.KeyType == dynamodb.KeyTypeHash {
					d.Set("hash_key", attribute.AttributeName)
				}

				if *attribute.KeyType == dynamodb.KeyTypeRange {
					d.Set("range_key", attribute.AttributeName)
				}
			}

			nonKeyAttrs := make([]string, 0, len(gsiObject.Projection.NonKeyAttributes))
			for _, nonKeyAttr := range gsiObject.Projection.NonKeyAttributes {
				nonKeyAttrs = append(nonKeyAttrs, *nonKeyAttr)
			}
			d.Set("non_key_attributes", nonKeyAttrs)

			break
		}
	}

	return nil
}
