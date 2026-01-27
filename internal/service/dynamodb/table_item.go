// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_table_item", name="Table Item")
func resourceTableItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableItemCreate,
		ReadWithoutTimeout:   resourceTableItemRead,
		UpdateWithoutTimeout: resourceTableItemUpdate,
		DeleteWithoutTimeout: resourceTableItemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTableItemImportState,
		},

		Schema: map[string]*schema.Schema{
			"hash_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"item": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validateTableItem,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
			},
			"range_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func validateTableItem(v any, k string) (ws []string, errors []error) {
	_, err := expandTableItemAttributes(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("Invalid format of %q: %w", k, err))
	}
	return
}

func resourceTableItemCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	attributes, err := expandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tableName := d.Get(names.AttrTableName).(string)
	hashKey := d.Get("hash_key").(string)
	input := &dynamodb.PutItemInput{
		// Explode if item exists. We didn't create it.
		ConditionExpression:      aws.String("attribute_not_exists(#hk)"),
		ExpressionAttributeNames: map[string]string{"#hk": hashKey},
		Item:                     attributes,
		TableName:                aws.String(tableName),
	}

	_, err = conn.PutItem(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Table (%s) Item: %s", tableName, err)
	}

	d.SetId(tableItemCreateResourceID(tableName, hashKey, d.Get("range_key").(string), attributes))

	return append(diags, resourceTableItemRead(ctx, d, meta)...)
}

func resourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrTableName).(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	attributes, err := expandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	key := expandTableItemQueryKey(attributes, hashKey, rangeKey)
	item, err := findTableItemByTwoPartKey(ctx, conn, tableName, key)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	// The record exists, now test if it differs from what is desired
	if !reflect.DeepEqual(item, attributes) {
		itemAttrs, err := flattenTableItemAttributes(item)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("item", itemAttrs)
		d.SetId(tableItemCreateResourceID(tableName, hashKey, rangeKey, item))
	}

	return diags
}

func resourceTableItemUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	if d.HasChange("item") {
		tableName := d.Get(names.AttrTableName).(string)
		hashKey := d.Get("hash_key").(string)
		rangeKey := d.Get("range_key").(string)

		oldItem, newItem := d.GetChange("item")

		attributes, err := expandTableItemAttributes(newItem.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		newQueryKey := expandTableItemQueryKey(attributes, hashKey, rangeKey)

		updates := map[string]awstypes.AttributeValueUpdate{}
		for key, value := range attributes {
			// Hash keys and range keys are not updatable, so we'll basically create
			// a new record and delete the old one below
			if key == hashKey || key == rangeKey {
				continue
			}
			updates[key] = awstypes.AttributeValueUpdate{
				Action: awstypes.AttributeActionPut,
				Value:  value,
			}
		}

		oldAttributes, err := expandTableItemAttributes(oldItem.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		for k := range oldAttributes {
			if k == hashKey || k == rangeKey {
				continue
			}
			if _, ok := attributes[k]; !ok {
				updates[k] = awstypes.AttributeValueUpdate{
					Action: awstypes.AttributeActionDelete,
				}
			}
		}

		input := &dynamodb.UpdateItemInput{
			AttributeUpdates: updates,
			Key:              newQueryKey,
			TableName:        aws.String(tableName),
		}

		_, err = conn.UpdateItem(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): %s", d.Id(), err)
		}

		// New record is created via UpdateItem in case we're changing hash key
		// so we need to get rid of the old one
		oldQueryKey := expandTableItemQueryKey(oldAttributes, hashKey, rangeKey)
		if !reflect.DeepEqual(oldQueryKey, newQueryKey) {
			input := &dynamodb.DeleteItemInput{
				Key:       oldQueryKey,
				TableName: aws.String(tableName),
			}

			_, err := conn.DeleteItem(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating DynamoDB Table Item (%s): removing old record: %s", d.Id(), err)
			}
		}

		d.SetId(tableItemCreateResourceID(tableName, hashKey, rangeKey, attributes))
	}

	return append(diags, resourceTableItemRead(ctx, d, meta)...)
}

func resourceTableItemDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	attributes, err := expandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	queryKey := expandTableItemQueryKey(attributes, hashKey, rangeKey)

	input := dynamodb.DeleteItemInput{
		Key:       queryKey,
		TableName: aws.String(d.Get(names.AttrTableName).(string)),
	}
	_, err = conn.DeleteItem(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	return diags
}

func tableItemCreateResourceID(tableName string, hashKey string, rangeKey string, attrs map[string]awstypes.AttributeValue) string {
	id := []string{tableName, hashKey}

	if v, ok := attrs[hashKey]; ok {
		switch v := v.(type) {
		case *awstypes.AttributeValueMemberB:
			id = append(id, inttypes.Base64EncodeOnce(v.Value))
		case *awstypes.AttributeValueMemberN:
			id = append(id, v.Value)
		case *awstypes.AttributeValueMemberS:
			id = append(id, v.Value)
		}
	}

	if v, ok := attrs[rangeKey]; ok && rangeKey != "" {
		switch v := v.(type) {
		case *awstypes.AttributeValueMemberB:
			id = append(id, inttypes.Base64EncodeOnce(v.Value))
		case *awstypes.AttributeValueMemberN:
			id = append(id, v.Value)
		case *awstypes.AttributeValueMemberS:
			id = append(id, v.Value)
		}
	}

	return strings.Join(id, "|")
}

func findTableItemByTwoPartKey(ctx context.Context, conn *dynamodb.Client, tableName string, key map[string]awstypes.AttributeValue) (map[string]awstypes.AttributeValue, error) {
	input := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key:            key,
		TableName:      aws.String(tableName),
	}

	return findTableItem(ctx, conn, input)
}

func findTableItem(ctx context.Context, conn *dynamodb.Client, input *dynamodb.GetItemInput) (map[string]awstypes.AttributeValue, error) {
	output, err := conn.GetItem(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Item == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Item, nil
}

func expandTableItemQueryKey(attrs map[string]awstypes.AttributeValue, hashKey, rangeKey string) map[string]awstypes.AttributeValue {
	queryKey := map[string]awstypes.AttributeValue{
		hashKey: attrs[hashKey],
	}
	if rangeKey != "" {
		queryKey[rangeKey] = attrs[rangeKey]
	}

	return queryKey
}

func createTableItemKeyAttr(attrTypes map[string]awstypes.ScalarAttributeType, name, value string) (awstypes.AttributeValue, error) {
	attrType, ok := attrTypes[name]
	if !ok {
		return nil, fmt.Errorf("key %s not found in attribute definitions", name)
	}
	switch attrType {
	case awstypes.ScalarAttributeTypeS:
		return &awstypes.AttributeValueMemberS{Value: value}, nil
	case awstypes.ScalarAttributeTypeN:
		return &awstypes.AttributeValueMemberN{Value: value}, nil
	case awstypes.ScalarAttributeTypeB:
		data, err := itypes.Base64Decode(value)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 value for binary attribute %s: %s", name, err)
		}
		return &awstypes.AttributeValueMemberB{Value: data}, nil
	default:
		return nil, fmt.Errorf("unsupported attribute type: %s", attrType)
	}
}

func resourceTableItemImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	idParts := strings.Split(d.Id(), "|")
	if len(idParts) < 3 || len(idParts) > 4 {
		return nil, fmt.Errorf("unexpected format for import ID (%s), expected tableName|hashKeyName|hashKeyValue[|rangeKeyValue]", d.Id())
	}

	tableName := idParts[0]
	hashKey := idParts[1]
	hashValue := idParts[2]
	var rangeValue string
	if len(idParts) == 4 {
		rangeValue = idParts[3]
	}

	output, err := conn.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("describing table %s: %s", tableName, err)
	}

	var rangeKey string
	if rangeValue != "" {
		var found bool
		for _, elem := range output.Table.KeySchema {
			if aws.ToString(elem.AttributeName) != hashKey && elem.KeyType == awstypes.KeyTypeRange {
				rangeKey = aws.ToString(elem.AttributeName)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("import ID contains range key value but table %s does not have a range key", tableName)
		}
	}

	attrTypes := map[string]awstypes.ScalarAttributeType{}
	for _, v := range output.Table.AttributeDefinitions {
		attrTypes[aws.ToString(v.AttributeName)] = v.AttributeType
	}

	key := map[string]awstypes.AttributeValue{}
	key[hashKey], err = createTableItemKeyAttr(attrTypes, hashKey, hashValue)
	if err != nil {
		return nil, err
	}
	if rangeValue != "" {
		key[rangeKey], err = createTableItemKeyAttr(attrTypes, rangeKey, rangeValue)
		if err != nil {
			return nil, err
		}
	}

	item, err := findTableItemByTwoPartKey(ctx, conn, tableName, key)
	if err != nil {
		return nil, fmt.Errorf("reading DynamoDB Table Item: %s: %s", d.Id(), err)
	}
	itemAttrs, err := flattenTableItemAttributes(item)
	if err != nil {
		return nil, fmt.Errorf("flattening item attributes: %s", err)
	}

	d.Set(names.AttrTableName, tableName)
	d.Set("hash_key", hashKey)
	if rangeKey != "" {
		d.Set("range_key", rangeKey)
	}
	d.Set("item", itemAttrs)
	d.SetId(tableItemCreateResourceID(tableName, hashKey, rangeKey, item))

	return []*schema.ResourceData{d}, nil
}
