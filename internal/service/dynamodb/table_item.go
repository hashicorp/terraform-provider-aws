// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_dynamodb_table_item")
func ResourceTableItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableItemCreate,
		ReadWithoutTimeout:   resourceTableItemRead,
		UpdateWithoutTimeout: resourceTableItemUpdate,
		DeleteWithoutTimeout: resourceTableItemDelete,

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

func BuildExpressionAttributeNames(attrs map[string]*dynamodb.AttributeValue) map[string]*string {
	names := map[string]*string{}

	for key := range attrs {
		names["#a_"+cleanKeyName(key)] = aws.String(key)
	}

	log.Printf("[DEBUG] ExpressionAttributeNames: %+v", names)
	return names
}

func cleanKeyName(key string) string {
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		log.Printf("[ERROR] clean keyname errored %v", err)
	}
	return reg.ReplaceAllString(key, "")
}

func BuildProjectionExpression(attrs map[string]*dynamodb.AttributeValue) *string {
	keys := []string{}

	for key := range attrs {
		keys = append(keys, cleanKeyName(key))
	}
	log.Printf("[DEBUG] ProjectionExpressions: %+v", strings.Join(keys, ", #a_"))
	return aws.String("#a_" + strings.Join(keys, ", #a_"))
}

func buildTableItemID(tableName string, hashKey string, rangeKey string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName, hashKey}

	if hashVal, ok := attrs[hashKey]; ok {
		id = append(id, verify.Base64Encode(hashVal.B))
		id = append(id, aws.StringValue(hashVal.S))
		id = append(id, aws.StringValue(hashVal.N))
	}
	if rangeVal, ok := attrs[rangeKey]; ok && rangeKey != "" {
		id = append(id, rangeKey, verify.Base64Encode(rangeVal.B))
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
