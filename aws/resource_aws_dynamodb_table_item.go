package aws

import (
	"fmt"
	"log"
	strings "strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"

	"bytes"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	reflect "reflect"
)

// A number of these are marked as computed because if you don't
// provide a value, DynamoDB will provide you with defaults (which are the
// default values specified below)
func resourceAwsDynamoDbTableItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableItemCreate,
		Read:   resourceAwsDynamoDbTableItemRead,
		Update: resourceAwsDynamoDbTableItemUpdate,
		Delete: resourceAwsDynamoDbTableItemDelete,
		Schema: map[string]*schema.Schema{
			"table_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"hash_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"item": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"range_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"query_key": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"range_value": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"hash_value": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"consumed_capacity": &schema.Schema{
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"last_modified": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDynamoDbTableItemCreate(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)

	log.Printf("[DEBUG] DynamoDB item create: %s", tableName)

	item := d.Get("item").(string)

	var av map[string]*dynamodb.AttributeValue

	avDec := json.NewDecoder(strings.NewReader(item))

	if err := avDec.Decode(&av); err != nil {
		return fmt.Errorf("Error deserializing DynamoDB item JSON: %s", err)
	}

	exists := false
	req := &dynamodb.PutItemInput{
		Item: av,
		Expected: map[string]*dynamodb.ExpectedAttributeValue{
			hashKey: {
				// Explode if item exists. We didn't create it.
				Exists: &exists,
			},
		},
		TableName: &tableName,
	}

	id := getId(tableName, hashKey, rangeKey, av)
	err := retryLoop(func() error {
		_, err := dynamodbconn.PutItem(req)

		return err
	}, fmt.Sprintf("creating DynamoDB table item '%s'", id))

	if err != nil {
		return err
	}

	setQueryKey(d, av, item, hashKey, rangeKey)

	d.SetId(id)

	return resourceAwsDynamoDbTableItemRead(d, meta)
}

func getId(tableName string, hashKey string, rangeKey string, av map[string]*dynamodb.AttributeValue) string {
	hashVal := av[hashKey]

	id := []string{
		tableName,
		hashKey,
		base64Encode(hashVal.B),
	}

	if hashVal.S != nil {
		id = append(id, *hashVal.S)
	} else {
		id = append(id, "")
	}
	if hashVal.N != nil {
		id = append(id, *hashVal.N)
	} else {
		id = append(id, "")
	}
	if rangeKey != "" {
		rangeVal := av[rangeKey]

		id = append(id,
			rangeKey,
			base64Encode(rangeVal.B),
		)

		if rangeVal.S != nil {
			id = append(id, *rangeVal.S)
		} else {
			id = append(id, "")
		}

		if rangeVal.N != nil {
			id = append(id, *rangeVal.N)
		} else {
			id = append(id, "")
		}

	}

	return strings.Join(id, "|")
}

func resourceAwsDynamoDbTableItemUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating DynamoDB table %s", d.Id())
	dynamodbconn := meta.(*AWSClient).dynamodbconn

	if d.HasChange("item") {
		o, n := d.GetChange("item")

		tableName := d.Get("table_name").(string)
		hashKey := d.Get("hash_key").(string)
		rangeKey := d.Get("range_key").(string)

		newJson := n.(string)

		var newItem map[string]*dynamodb.AttributeValue

		newDec := json.NewDecoder(strings.NewReader(newJson))
		if err := newDec.Decode(&newItem); err != nil {
			return fmt.Errorf("Error deserializing DynamoDB item JSON: %s", err)
		}

		newQueryKey := getQueryKey(newItem, hashKey, rangeKey)

		updates := map[string]*dynamodb.AttributeValueUpdate{}

		for k, v := range newItem {
			// We shouldn't update the key values
			skip := false
			for qk := range newQueryKey {
				if skip = (qk == k); skip {
					break
				}
			}
			if skip {
				continue
			}

			action := "PUT"
			updates[k] = &dynamodb.AttributeValueUpdate{
				Action: &action,
				Value:  v,
			}
		}

		req := dynamodb.UpdateItemInput{
			AttributeUpdates: updates,
			TableName:        &tableName,
			Key:              newQueryKey,
		}

		err := retryLoop(func() error {
			_, err := dynamodbconn.UpdateItem(&req)

			return err
		}, "updating DynamoDB table item '%s'")

		if err != nil {
			return err
		}

		// If we finished successfully, delete the old record if the query key is different
		oldJson := o.(string)

		var oldItem map[string]*dynamodb.AttributeValue

		oldDec := json.NewDecoder(strings.NewReader(oldJson))
		if err := oldDec.Decode(&oldItem); err != nil {
			return fmt.Errorf("Error deserializing DynamoDB item JSON: %s", err)
		}

		oldQueryKey := getQueryKey(oldItem, hashKey, rangeKey)

		id := getId(tableName, hashKey, rangeKey, newItem)

		if !reflect.DeepEqual(oldQueryKey, newQueryKey) {
			req := dynamodb.DeleteItemInput{
				Key:       oldQueryKey,
				TableName: &tableName,
			}

			err := retryLoop(func() error {
				_, err := dynamodbconn.DeleteItem(&req)
				return err
			}, fmt.Sprintf("deleting old DynamoDB item '%s'", id))

			if err != nil {
				return err
			}
		}

		setQueryKey(d, newItem, newJson, hashKey, rangeKey)

		d.SetId(id)
	}

	return resourceAwsDynamoDbTableItemRead(d, meta)
}

func getQueryKey(av map[string]*dynamodb.AttributeValue, hashKey string, rangeKey string) map[string]*dynamodb.AttributeValue {
	qk := map[string]*dynamodb.AttributeValue{}

	flen := 1
	if rangeKey != "" {
		flen = 2
	}

	for k, v := range av {
		if k == hashKey || k == rangeKey {
			qk[k] = v
		}
		if len(qk) == flen {
			break
		}
	}

	return qk
}

func setQueryKey(d *schema.ResourceData, av map[string]*dynamodb.AttributeValue, item string, hashKey string, rangeKey string) {
	var itemRaw map[string]json.RawMessage

	hashDec := json.NewDecoder(strings.NewReader(item))
	hashDec.Decode(&itemRaw)

	keyRaw := map[string]json.RawMessage{}

	hashRaw := itemRaw[hashKey]

	keyRaw[hashKey] = hashRaw

	hashBytes, _ := hashRaw.MarshalJSON()
	d.Set("hash_value", string(hashBytes))

	if rangeKey != "" {
		rangeRaw := itemRaw[rangeKey]

		keyRaw[rangeKey] = rangeRaw

		rangeBytes, _ := rangeRaw.MarshalJSON()
		d.Set("range_value", string(rangeBytes))
	}

	queryKeyBuf := bytes.NewBufferString("")
	queryKeyEnc := json.NewEncoder(queryKeyBuf)
	queryKeyEnc.Encode(keyRaw)

	d.Set("query_key", queryKeyBuf.String())
}

func retryLoop(action func() error, actionDetails string) error {
	attemptCount := 1
	for attemptCount <= DYNAMODB_MAX_THROTTLE_RETRIES {
		err := action()

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				switch code := awsErr.Code(); code {
				case "ThrottlingException":
					log.Printf("[DEBUG] Attempt %d/%d: Sleeping for a bit to throttle back %s", attemptCount, DYNAMODB_MAX_THROTTLE_RETRIES, actionDetails)
					time.Sleep(DYNAMODB_THROTTLE_SLEEP)
					attemptCount += 1
				case "LimitExceededException":
					// If we're at resource capacity, error out without retry
					if strings.Contains(awsErr.Message(), "Subscriber limit exceeded:") {
						return fmt.Errorf("AWS Error %s: %s", actionDetails, err)
					}
					log.Printf("[DEBUG] Limit on concurrency of %s, sleeping for a bit", actionDetails)
					time.Sleep(DYNAMODB_LIMIT_EXCEEDED_SLEEP)
					attemptCount += 1
				default:
					// Some other non-retryable exception occurred
					return fmt.Errorf("AWS Error %s: %s", actionDetails, err)
				}
			} else {
				// Non-AWS exception occurred, give up
				return fmt.Errorf("Error %s: %s", actionDetails, err)
			}
		} else {
			return nil
		}
	}

	// Too many throttling events occurred, give up
	return fmt.Errorf("Failed %s after %d attempts", actionDetails, attemptCount)
}

func resourceAwsDynamoDbTableItemRead(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn
	log.Printf("[DEBUG] Loading data for DynamoDB table item '%s'", d.Id())

	tableName := d.Get("table_name").(string)

	// The record exists, now test if it differs from what is desired
	item := d.Get("item").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)

	var av map[string]*dynamodb.AttributeValue
	itemDec := json.NewDecoder(strings.NewReader(item))
	itemDec.Decode(&av)

	itemAttributes := []string{}
	for k := range av {
		itemAttributes = append(itemAttributes, k)
	}

	queryKey := getQueryKey(av, hashKey, rangeKey)

	expressionAttributeNames := map[string]*string{}
	projection := "#a_" + strings.Join(itemAttributes, ", #a_")

	for _, v := range itemAttributes {
		w := v
		expressionAttributeNames["#a_"+v] = &w
	}

	req := dynamodb.GetItemInput{
		TableName:                &tableName,
		Key:                      queryKey,
		ProjectionExpression:     &projection,
		ExpressionAttributeNames: expressionAttributeNames,
	}

	result, err := dynamodbconn.GetItem(&req)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Dynamodb Table Item (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving DynamoDB table item: %s %s", err, req)
	}

	// The record exists, now test if it differs from what is desired
	if result.Item != nil && !reflect.DeepEqual(result.Item, av) {
		buf := bytes.NewBufferString("")
		enc := json.NewEncoder(buf)
		enc.Encode(result.Item)

		var itemRaw map[string]map[string]interface{}

		// Reserialize so we get rid of the nulls
		dec := json.NewDecoder(strings.NewReader(buf.String()))
		dec.Decode(&itemRaw)

		for _, val := range itemRaw {
			for typeName, typeVal := range val {
				if typeVal == nil {
					delete(val, typeName)
				}
			}
		}

		rawBuf := bytes.NewBufferString("")
		rawEnc := json.NewEncoder(rawBuf)
		rawEnc.Encode(itemRaw)

		d.Set("item", rawBuf.String())

		id := getId(tableName, hashKey, rangeKey, result.Item)
		d.SetId(id)
	} else if result.Item == nil {
		d.SetId("")
	}

	d.Set("consumed_capacity", result.ConsumedCapacity)

	return nil
}

func resourceAwsDynamoDbTableItemDelete(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)

	item := d.Get("item").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)

	var av map[string]*dynamodb.AttributeValue
	itemDec := json.NewDecoder(strings.NewReader(item))
	itemDec.Decode(&av)

	queryKey := getQueryKey(av, hashKey, rangeKey)

	req := dynamodb.DeleteItemInput{
		Key:       queryKey,
		TableName: &tableName,
	}

	err := retryLoop(func() error {
		_, err := dynamodbconn.DeleteItem(&req)

		return err
	}, fmt.Sprintf("deleting DynamoDB table item '%s'", d.Id()))

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
