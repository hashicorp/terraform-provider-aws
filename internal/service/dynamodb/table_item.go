// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_dynamodb_table_item")
func ResourceTableItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableItemCreate,
		ReadWithoutTimeout:   resourceTableItemRead,
		UpdateWithoutTimeout: resourceTableItemUpdate,
		DeleteWithoutTimeout: resourceTableItemDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTableItemImport,
		},

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
			"item": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validateTableItem,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
			},
		},
	}
}

func validateTableItem(v interface{}, k string) (ws []string, errors []error) {
	_, err := ExpandTableItemAttributes(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("Invalid format of %q: %s", k, err))
	}
	return
}

func resourceTableItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	item := d.Get("item").(string)
	attributes, err := ExpandTableItemAttributes(item)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Table Item: %s", err)
	}

	log.Printf("[DEBUG] DynamoDB item create: %s", tableName)

	_, err = conn.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: attributes,
		// Explode if item exists. We didn't create it.
		ConditionExpression:      aws.String("attribute_not_exists(#hk)"),
		ExpressionAttributeNames: aws.StringMap(map[string]string{"#hk": hashKey}),
		TableName:                aws.String(tableName),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Table Item: %s", err)
	}

	rangeKey := d.Get("range_key").(string)
	id := buildTableItemID(tableName, hashKey, rangeKey, attributes)

	d.SetId(id)

	return append(diags, resourceTableItemRead(ctx, d, meta)...)
}

func resourceTableItemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Updating DynamoDB table %s", d.Id())
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	if d.HasChange("item") {
		tableName := d.Get("table_name").(string)
		hashKey := d.Get("hash_key").(string)
		rangeKey := d.Get("range_key").(string)

		oldItem, newItem := d.GetChange("item")

		attributes, err := ExpandTableItemAttributes(newItem.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): %s", d.Id(), err)
		}
		newQueryKey := BuildTableItemQueryKey(attributes, hashKey, rangeKey)

		updates := map[string]*dynamodb.AttributeValueUpdate{}
		for key, value := range attributes {
			// Hash keys and range keys are not updatable, so we'll basically create
			// a new record and delete the old one below
			if key == hashKey || key == rangeKey {
				continue
			}
			updates[key] = &dynamodb.AttributeValueUpdate{
				Action: aws.String(dynamodb.AttributeActionPut),
				Value:  value,
			}
		}

		oldAttributes, err := ExpandTableItemAttributes(oldItem.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): %s", d.Id(), err)
		}

		for k := range oldAttributes {
			if k == hashKey || k == rangeKey {
				continue
			}
			if _, ok := attributes[k]; !ok {
				updates[k] = &dynamodb.AttributeValueUpdate{
					Action: aws.String(dynamodb.AttributeActionDelete),
				}
			}
		}

		_, err = conn.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
			AttributeUpdates: updates,
			TableName:        aws.String(tableName),
			Key:              newQueryKey,
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): %s", d.Id(), err)
		}

		// New record is created via UpdateItem in case we're changing hash key
		// so we need to get rid of the old one
		oldQueryKey := BuildTableItemQueryKey(oldAttributes, hashKey, rangeKey)
		if !reflect.DeepEqual(oldQueryKey, newQueryKey) {
			log.Printf("[DEBUG] Deleting old record: %#v", oldQueryKey)
			_, err := conn.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
				Key:       oldQueryKey,
				TableName: aws.String(tableName),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): removing old record: %s", d.Id(), err)
			}
		}

		id := buildTableItemID(tableName, hashKey, rangeKey, attributes)
		d.SetId(id)
	}

	return append(diags, resourceTableItemRead(ctx, d, meta)...)
}

func resourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	log.Printf("[DEBUG] Loading data for DynamoDB table item '%s'", d.Id())

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	attributes, err := ExpandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	key := BuildTableItemQueryKey(attributes, hashKey, rangeKey)
	result, err := FindTableItem(ctx, conn, tableName, key)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	// The record exists, now test if it differs from what is desired
	if !reflect.DeepEqual(result.Item, attributes) {
		itemAttrs, err := flattenTableItemAttributes(result.Item)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", d.Id(), err)
		}
		d.Set("item", itemAttrs)
		id := buildTableItemID(tableName, hashKey, rangeKey, result.Item)
		d.SetId(id)
	}

	return diags
}

func resourceTableItemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	attributes, err := ExpandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Table Item (%s): %s", d.Id(), err)
	}
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	queryKey := BuildTableItemQueryKey(attributes, hashKey, rangeKey)

	_, err = conn.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		Key:       queryKey,
		TableName: aws.String(d.Get("table_name").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	return diags
}

func parseTableItemQueryKey(d *schema.ResourceData, keyType string, keyName string, keyValue string, tableDescription *dynamodb.TableDescription, attrs map[string]*dynamodb.AttributeValue) error {
	if keyValue == "" {
		return fmt.Errorf("No value given for %s %s", keyType, keyName)
	}

	var value *dynamodb.AttributeValue
	found := false

	// Find the matching attribute definition and construct an appropriate attribute value based on the type
	for _, attr := range tableDescription.AttributeDefinitions {
		if *attr.AttributeName == keyName {
			found = true
			switch *attr.AttributeType {
			case "B":
				data, err := base64.StdEncoding.DecodeString(keyValue)
				if err != nil {
					return err
				}
				value = &dynamodb.AttributeValue{
					B: data,
				}
			case "S":
				value = &dynamodb.AttributeValue{
					S: aws.String(keyValue),
				}
			case "N":
				value = &dynamodb.AttributeValue{
					N: aws.String(keyValue),
				}
			default:
				return fmt.Errorf("%s %s has invalid type %s", keyType, keyName, *attr.AttributeType)
			}
		}
	}

	// Set both the resource attribute and the attribute query parameter
	if !found {
		return fmt.Errorf("%s %s not found in table", keyType, keyName)
	}
	d.Set(keyType, keyName)
	attrs[keyName] = value

	return nil
}

func resourceTableItemImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	// Parse given id string as either a pipe-delimited string or json
	var id []string
	if strings.HasPrefix(d.Id(), "[") {
		err := json.Unmarshal([]byte(d.Id()), &id)
		if err != nil {
			return nil, err
		}
	} else {
		id = strings.Split(d.Id(), "|")
	}

	// Check for proper number of array elements
	if len(id) == 2 {
		id = append(id, "")
	}
	if len(id) != 3 || id[0] == "" || id[1] == "" {
		return nil, errors.New("Invalid id, must be of the form table_name|hash_key_value[|range_key_value] or json [ \"table_name\", \"hash_key_value\", \"range_key_value\" ]")
	}

	// Initialize table query parameters
	tableName := id[0]
	hashKey := ""
	rangeKey := ""
	hashKeyValueString := id[1]
	rangeKeyValueString := id[2]
	params := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),
		Key:            map[string]*dynamodb.AttributeValue{},
	}

	// Query table description to determine its hash/range key attributes
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)
	tableResult, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}
	tableDescription := tableResult.Table

	// Build attribute query using given hash/range key values
	for _, key := range tableDescription.KeySchema {
		switch *key.KeyType {
		case "HASH":
			hashKey = *key.AttributeName
			err := parseTableItemQueryKey(d, "hash_key", hashKey, hashKeyValueString, tableDescription, params.Key)
			if err != nil {
				return nil, err
			}
		case "RANGE":
			rangeKey = *key.AttributeName
			err := parseTableItemQueryKey(d, "range_key", rangeKey, rangeKeyValueString, tableDescription, params.Key)
			if err != nil {
				return nil, err
			}
		}
	}

	// Error if we were given a range key value but the table has no range key
	if rangeKey == "" && rangeKeyValueString != "" {
		return nil, fmt.Errorf("Table %s has no range key but a range key value was given", tableName)
	}

	// Query table for matching record
	result, err := conn.GetItem(params)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("No item matching %s found to import", d.Id())
	}
	itemAttrs, err := flattenTableItemAttributes(result.Item)
	if err != nil {
		return nil, err
	}

	// Set required resource attributes
	d.Set("table_name", tableName)
	d.Set("hash_key", hashKey)
	if rangeKey != "" {
		d.Set("range_key", rangeKey)
	}
	d.Set("item", itemAttrs)

	// Always set id to canonical format
	d.SetId(buildTableItemID(tableName, hashKey, rangeKey, params.Key))

	return []*schema.ResourceData{d}, nil
}

// Helpers

func FindTableItem(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, key map[string]*dynamodb.AttributeValue) (*dynamodb.GetItemOutput, error) {
	in := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),
		Key:            key,
	}

	out, err := conn.GetItemWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Item == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func buildTableItemID(tableName string, hashKey string, rangeKey string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName, hashKey}

	if hashVal, ok := attrs[hashKey]; ok {
		id = append(id, itypes.Base64EncodeOnce(hashVal.B))
		id = append(id, aws.StringValue(hashVal.S))
		id = append(id, aws.StringValue(hashVal.N))
	}
	if rangeVal, ok := attrs[rangeKey]; ok && rangeKey != "" {
		id = append(id, rangeKey, itypes.Base64EncodeOnce(rangeVal.B))
		id = append(id, aws.StringValue(rangeVal.S))
		id = append(id, aws.StringValue(rangeVal.N))
	}
	return strings.Join(id, "|")
}

func BuildTableItemQueryKey(attrs map[string]*dynamodb.AttributeValue, hashKey string, rangeKey string) map[string]*dynamodb.AttributeValue {
	queryKey := map[string]*dynamodb.AttributeValue{
		hashKey: attrs[hashKey],
	}
	if rangeKey != "" {
		queryKey[rangeKey] = attrs[rangeKey]
	}
	return queryKey
}
