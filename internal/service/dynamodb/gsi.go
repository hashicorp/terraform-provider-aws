package dynamodb

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAwsDynamoDbTableGsi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableGsiCreate,
		Read:   resourceAwsDynamoDbTableGsiRead,
		Update: resourceAwsDynamoDbTableGsiUpdate,
		Delete: resourceAwsDynamoDbTableGsiDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute), // provisioned throughput changes only
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

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
				Optional: true,
				ForceNew: true,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dynamodb.ProjectionType_Values(), false),
			},
			"non_key_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"billing_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dynamodb.BillingModeProvisioned,
				ValidateFunc: validation.StringInSlice(dynamodb.BillingMode_Values(), false),
			},
			"attribute": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
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
					return create.StringHashcode(buf.String())
				},
			},
		},
	}
}

func resourceAwsDynamoDbTableGsiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	tableName := d.Get("table_name").(string)
	indexName := d.Get("name").(string)
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

	projection := &dynamodb.Projection{
		ProjectionType: aws.String(d.Get("projection_type").(string)),
	}

	if v, ok := d.Get("non_key_attributes").(*schema.Set); ok {
		projection.NonKeyAttributes = flex.ExpandStringList(v.List())
	}

	capacityMap := map[string]interface{}{
		"write_capacity": d.Get("write_capacity"),
		"read_capacity":  d.Get("read_capacity"),
	}
	billingMode := d.Get("billing_mode").(string)

	createOp := &dynamodb.GlobalSecondaryIndexUpdate{
		Create: &dynamodb.CreateGlobalSecondaryIndexAction{
			IndexName:             aws.String(indexName),
			KeySchema:             expandDynamoDbKeySchema(keySchemaMap),
			Projection:            projection,
			ProvisionedThroughput: expandDynamoDbProvisionedThroughput(capacityMap, billingMode),
		},
	}

	req.GlobalSecondaryIndexUpdates = []*dynamodb.GlobalSecondaryIndexUpdate{createOp}
	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	var output *dynamodb.UpdateTableOutput
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var err error
		output, err = conn.UpdateTable(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		return fmt.Errorf(`Updating table timed out: %s`, err)
	}
	if err != nil {
		return err
	}

	gsiDescription := findDynamoDbGsi(&output.TableDescription.GlobalSecondaryIndexes, indexName)
	d.SetId(
		aws.StringValue(gsiDescription.IndexName),
	)

	_, err = waitDynamoDBGSIActive(conn, d.Get("table_name").(string), d.Id())
	return err
}

func resourceAwsDynamoDbTableGsiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	capacityMap := map[string]interface{}{
		"write_capacity": d.Get("write_capacity"),
		"read_capacity":  d.Get("read_capacity"),
	}
	billingMode := d.Get("billing_mode").(string)
	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
					IndexName:             aws.String(d.Id()),
					ProvisionedThroughput: expandDynamoDbProvisionedThroughput(capacityMap, billingMode),
				},
			},
		},
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		var err error
		_, err = conn.UpdateTable(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		return fmt.Errorf(`Updating table timed out: %s`, err)
	}
	if err != nil {
		return err
	}

	_, err = waitDynamoDBGSIActive(conn, d.Get("table_name").(string), d.Id())
	return err
}

func resourceAwsDynamoDbTableGsiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn
	tableName := d.Get("table_name").(string)
	d.Set("table_name", tableName)

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return err
	}

	gsi := findDynamoDbGsi(&result.Table.GlobalSecondaryIndexes, d.Id())
	d.Set("write_capacity", gsi.ProvisionedThroughput.WriteCapacityUnits)
	d.Set("read_capacity", gsi.ProvisionedThroughput.ReadCapacityUnits)
	d.Set("projection_type", gsi.Projection.ProjectionType)
	d.Set("non_key_attributes", gsi.Projection.NonKeyAttributes)

	gsiAttributeNames := make(map[string]struct{}, len(gsi.KeySchema))
	for _, attribute := range gsi.KeySchema {
		if aws.StringValue(attribute.KeyType) == dynamodb.KeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
			gsiAttributeNames[*attribute.AttributeName] = struct{}{}
		}

		if aws.StringValue(attribute.KeyType) == dynamodb.KeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
			gsiAttributeNames[*attribute.AttributeName] = struct{}{}
		}
	}
	attributes := []interface{}{}
	for _, attrdef := range result.Table.AttributeDefinitions {
		if _, ok := gsiAttributeNames[*attrdef.AttributeName]; ok {
			attribute := map[string]string{
				"name": *attrdef.AttributeName,
				"type": *attrdef.AttributeType,
			}
			attributes = append(attributes, attribute)
		}
	}
	d.Set("attribute", attributes)

	return err
}

func resourceAwsDynamoDbTableGsiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
					IndexName: aws.String(d.Id()),
				},
			},
		},
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.UpdateTable(req)
		if err != nil {
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		return fmt.Errorf(`Updating table timed out: %s`, err)
	}
	if err != nil {
		return err
	}

	err = waitDynamoDBGSIDeleted(conn, d.Get("table_name").(string), d.Id())
	return err
}

func findDynamoDbGsi(gsiList *[]*dynamodb.GlobalSecondaryIndexDescription, target string) *dynamodb.GlobalSecondaryIndexDescription {
	for _, gsiObject := range *gsiList {
		if aws.StringValue(gsiObject.IndexName) == target {
			return gsiObject
		}
	}
	return nil
}
