package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	updateExpressionSet    = "SET"
	updateExpressionRemove = "REMOVE"
)

func resourceAwsDynamoDbTableItemAttribute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableItemAttributeUpdate,
		Read:   resourceAwsDynamoDbTableItemAttributeRead,
		Update: resourceAwsDynamoDbTableItemAttributeUpdate,
		Delete: resourceAwsDynamoDbTableItemAttributeDelete,

		Schema: map[string]*schema.Schema{
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
				ForceNew: true,
				Optional: true,
			},
			"attribute_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"attribute_value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsDynamoDbTableItemAttributeDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsDynamoDbTableItemAttributeModify(updateExpressionRemove, d, meta)
}

func resourceAwsDynamoDbTableItemAttributeUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := resourceAwsDynamoDbTableItemAttributeModify(updateExpressionSet, d, meta); err == nil {
		return resourceAwsDynamoDbTableItemAttributeRead(d, meta)
	} else {
		return err
	}
}

func resourceAwsDynamoDbTableItemAttributeModify(action string, d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] %s DynamoDB table %s", action, d.Id())
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)

	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	attributeKey := d.Get("attribute_key").(string)
	attributeValue := d.Get("attribute_value").(string)

	hashKeyName, rangeKeyName, err := resourceAwsDynamoDbTableItemAttributeGetKeysInfo(conn, tableName)
	if err != nil {
		return err
	}

	updateItemInput := &dynamodb.UpdateItemInput{
		Key:       resourceAwsDynamoDbTableItemAttributeGetQueryKey(hashKeyName, hashKey, rangeKeyName, rangeKey),
		TableName: aws.String(tableName),
	}

	if d.IsNewResource() {
		updateItemInput.ConditionExpression = aws.String(fmt.Sprintf("attribute_not_exists(%s)", attributeKey))
	}

	if action == updateExpressionSet {
		updateItemInput.UpdateExpression = aws.String(fmt.Sprintf("%s %s = :v", updateExpressionSet, attributeKey))
		updateItemInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":v": {
				S: aws.String(attributeValue),
			},
		}
	} else if action == updateExpressionRemove {
		updateItemInput.UpdateExpression = aws.String(fmt.Sprintf("%s %s", updateExpressionRemove, attributeKey))
	}

	if _, err := conn.UpdateItem(updateItemInput); err != nil {
		return err
	}

	id := fmt.Sprintf("%s:%s:%s:%s", tableName, hashKey, rangeKey, attributeKey)
	d.SetId(id)

	return nil
}

func resourceAwsDynamoDbTableItemAttributeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	log.Printf("[DEBUG] Loading data for DynamoDB table item attribute '%s'", d.Id())

	idParts := strings.Split(d.Id(), ":")
	tableName, hashKey, rangeKey, attributeKey := idParts[0], idParts[1], idParts[2], idParts[3]

	hashKeyName, rangeKeyName, err := resourceAwsDynamoDbTableItemAttributeGetKeysInfo(conn, tableName)
	if err != nil {
		return err
	}

	result, err := conn.GetItem(&dynamodb.GetItemInput{
		ConsistentRead:       aws.Bool(true),
		Key:                  resourceAwsDynamoDbTableItemAttributeGetQueryKey(hashKeyName, hashKey, rangeKeyName, rangeKey),
		TableName:            aws.String(tableName),
		ProjectionExpression: aws.String(attributeKey),
	})
	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table Item (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
	}

	if result.Item == nil {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found", d.Id())
		d.SetId("")
		return nil
	}
	d.Set("table_name", tableName)
	d.Set("hash_key", hashKey)
	d.Set("range_key", rangeKey)
	d.Set("attribute_key", attributeKey)
	d.Set("attribute_value", result.Item[attributeKey].S)

	return nil
}

func resourceAwsDynamoDbTableItemAttributeGetKeysInfo(conn *dynamodb.DynamoDB, tableName string) (string, string, error) {
	var hashKeyName, rangeKeyName string
	if out, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}); err == nil {
		for _, key := range out.Table.KeySchema {
			if *key.KeyType == dynamodb.KeyTypeHash {
				hashKeyName = *key.AttributeName
			} else if *key.KeyType == dynamodb.KeyTypeRange {
				rangeKeyName = *key.AttributeName
			}
		}
	} else {
		return "", "", fmt.Errorf("Error descring table %s: %v", tableName, err)
	}

	return hashKeyName, rangeKeyName, nil
}

func resourceAwsDynamoDbTableItemAttributeGetQueryKey(hashKeyName string, hashKeyValue string, rangeKeyName string, rangeKeyValue string) map[string]*dynamodb.AttributeValue {
	queryKey := map[string]*dynamodb.AttributeValue{
		hashKeyName: &dynamodb.AttributeValue{S: aws.String(hashKeyValue)},
	}
	if rangeKeyValue != "" {
		queryKey[rangeKeyName] = &dynamodb.AttributeValue{S: aws.String(rangeKeyValue)}
	}
	return queryKey
}
