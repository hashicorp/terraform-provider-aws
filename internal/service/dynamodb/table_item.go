// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dynamodb

import (
	"context"
	"encoding/base64"
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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_table_item", name="Table Item")
// @IdentityAttribute("table_name")
// @IdentityAttribute("hash_key_value")
// @IdentityAttribute("range_key_value", optional="true")
// @ImportIDHandler("tableItemImportID")
// @MutableIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;awstypes;map[string]awstypes.AttributeValue")
// @Testing(importIgnore="item")
// @Testing(plannableImportAction="NoOp")
// @Testing(preIdentityVersion="v6.51.0")
func resourceTableItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableItemCreate,
		ReadWithoutTimeout:   resourceTableItemRead,
		UpdateWithoutTimeout: resourceTableItemUpdate,
		DeleteWithoutTimeout: resourceTableItemDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"hash_key": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"hash_key_value": {
					Type:     schema.TypeString,
					Computed: true,
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
				"range_key_value": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTableName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			}
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

	rangeKey := d.Get("range_key").(string)
	hkv, rkv := tableItemKeyValues(attributes, hashKey, rangeKey)
	d.SetId(tableItemCreateResourceID(tableName, hkv, rkv))

	return append(diags, resourceTableItemRead(ctx, d, meta)...)
}

func resourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrTableName).(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	itemJSON := d.Get("item").(string)

	var key map[string]awstypes.AttributeValue

	if itemJSON != "" {
		// Normal path: derive the key from the `item` JSON in state.
		attrs, err := expandTableItemAttributes(itemJSON)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		key = expandTableItemQueryKey(attrs, hashKey, rangeKey)
	} else {
		// Post-import path: state has identity attributes but no `item`. Recover the key names
		// and AttributeValue types from the table's KeySchema/AttributeDefinitions.
		recoveredKey, recoveredHashKey, recoveredRangeKey, err := tableItemKeyFromIdentity(ctx, conn, d, tableName)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		key = recoveredKey
		hashKey = recoveredHashKey
		rangeKey = recoveredRangeKey
	}

	item, err := findTableItemByTwoPartKey(ctx, conn, tableName, key)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTableItemFlatten(d, tableName, hashKey, rangeKey, item)...)
}

// resourceTableItemFlatten populates the resource data with the canonical
// representation of a DynamoDB table item. Used by both Read and the list
// resource so that resource state is set consistently between the two paths.
//
// The `item` attribute uses SuppressEquivalentJSONDiffs with DiffSuppressOnRefresh,
// so always setting it here is safe even when the API response is byte-different
// from the user-supplied JSON.
func resourceTableItemFlatten(d *schema.ResourceData, tableName, hashKey, rangeKey string, item map[string]awstypes.AttributeValue) diag.Diagnostics {
	var diags diag.Diagnostics

	hkv, rkv := tableItemKeyValues(item, hashKey, rangeKey)

	d.SetId(tableItemCreateResourceID(tableName, hkv, rkv))
	d.Set(names.AttrTableName, tableName)
	d.Set("hash_key", hashKey)
	if hkv != "" {
		d.Set("hash_key_value", hkv)
	}
	if rangeKey != "" {
		d.Set("range_key", rangeKey)
		if rkv != "" {
			d.Set("range_key_value", rkv)
		}
	}

	itemAttrs, err := flattenTableItemAttributes(item)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("item", itemAttrs)

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

		hkvNew, rkvNew := tableItemKeyValues(attributes, hashKey, rangeKey)
		d.SetId(tableItemCreateResourceID(tableName, hkvNew, rkvNew))
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

func tableItemCreateResourceID(tableName string, hashKeyValue string, rangeKeyValue string) string {
	parts := []string{tableName, hashKeyValue}
	if rangeKeyValue != "" {
		parts = append(parts, rangeKeyValue)
	}
	return strings.Join(parts, flex.ResourceIdSeparator)
}

// tableItemKeyValues extracts the canonical string representations of the hash
// key and (optional) range key values from a fully-expanded item. Empty strings
// are returned for missing or non-key-eligible attribute types.
func tableItemKeyValues(attrs map[string]awstypes.AttributeValue, hashKey, rangeKey string) (string, string) {
	hashKeyValue := attributeValueToString(attrs[hashKey])
	var rangeKeyValue string
	if rangeKey != "" {
		rangeKeyValue = attributeValueToString(attrs[rangeKey])
	}
	return hashKeyValue, rangeKeyValue
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

// attributeValueToString returns the canonical string form of a DynamoDB key
// AttributeValue for use in identity attributes and resource IDs. Binary values
// are encoded as standard base64. Returns the empty string if the value is
// unset or of an unsupported (non-key-eligible) type.
func attributeValueToString(v awstypes.AttributeValue) string {
	switch v := v.(type) {
	case *awstypes.AttributeValueMemberB:
		return inttypes.Base64EncodeOnce(v.Value)
	case *awstypes.AttributeValueMemberN:
		return v.Value
	case *awstypes.AttributeValueMemberS:
		return v.Value
	default:
		return ""
	}
}

// attributeValueFromString is the inverse of attributeValueToString. It uses
// the table's declared ScalarAttributeType to build a typed AttributeValue
// from the canonical string form. The type information cannot be recovered
// from the string alone, which is why the table's AttributeDefinitions are
// consulted during import.
func attributeValueFromString(s string, attrType awstypes.ScalarAttributeType) (awstypes.AttributeValue, error) {
	switch attrType {
	case awstypes.ScalarAttributeTypeB:
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decoding base64 binary key value: %w", err)
		}
		return &awstypes.AttributeValueMemberB{Value: data}, nil
	case awstypes.ScalarAttributeTypeN:
		return &awstypes.AttributeValueMemberN{Value: s}, nil
	case awstypes.ScalarAttributeTypeS:
		return &awstypes.AttributeValueMemberS{Value: s}, nil
	default:
		return nil, fmt.Errorf("unsupported key attribute type: %q", attrType)
	}
}

// tableItemKeyFromIdentity builds a DynamoDB GetItem key from the resource's
// identity attributes (hash_key_value, range_key_value) and the table's
// KeySchema/AttributeDefinitions. It also returns the discovered hash key name
// and range key name so callers can populate state.
func tableItemKeyFromIdentity(ctx context.Context, conn *dynamodb.Client, d *schema.ResourceData, tableName string) (map[string]awstypes.AttributeValue, string, string, error) {
	hashKeyValue := d.Get("hash_key_value").(string)
	rangeKeyValue := d.Get("range_key_value").(string)

	if hashKeyValue == "" {
		return nil, "", "", fmt.Errorf("hash_key_value must be set to import DynamoDB Table Item from %s", tableName)
	}

	table, err := findTableByName(ctx, conn, tableName)
	if err != nil {
		return nil, "", "", fmt.Errorf("describing DynamoDB Table (%s): %w", tableName, err)
	}

	var hashKeyName, rangeKeyName string
	for _, k := range table.KeySchema {
		switch k.KeyType {
		case awstypes.KeyTypeHash:
			hashKeyName = aws.ToString(k.AttributeName)
		case awstypes.KeyTypeRange:
			rangeKeyName = aws.ToString(k.AttributeName)
		}
	}

	attrTypes := make(map[string]awstypes.ScalarAttributeType, len(table.AttributeDefinitions))
	for _, ad := range table.AttributeDefinitions {
		attrTypes[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	hashKeyType, ok := attrTypes[hashKeyName]
	if !ok {
		return nil, "", "", fmt.Errorf("table %s missing AttributeDefinition for hash key %q", tableName, hashKeyName)
	}
	hashKeyAV, err := attributeValueFromString(hashKeyValue, hashKeyType)
	if err != nil {
		return nil, "", "", fmt.Errorf("hash_key_value: %w", err)
	}

	key := map[string]awstypes.AttributeValue{hashKeyName: hashKeyAV}

	switch {
	case rangeKeyName != "":
		if rangeKeyValue == "" {
			return nil, "", "", fmt.Errorf("table %s has range key %q but range_key_value is not set", tableName, rangeKeyName)
		}
		rangeKeyType, ok := attrTypes[rangeKeyName]
		if !ok {
			return nil, "", "", fmt.Errorf("table %s missing AttributeDefinition for range key %q", tableName, rangeKeyName)
		}
		rangeKeyAV, err := attributeValueFromString(rangeKeyValue, rangeKeyType)
		if err != nil {
			return nil, "", "", fmt.Errorf("range_key_value: %w", err)
		}
		key[rangeKeyName] = rangeKeyAV
	case rangeKeyValue != "":
		return nil, "", "", fmt.Errorf("table %s has no range key but range_key_value is set to %q", tableName, rangeKeyValue)
	}

	return key, hashKeyName, rangeKeyName, nil
}

var _ inttypes.SDKv2ImportID = tableItemImportID{}

type tableItemImportID struct{}

// Create assembles a comma-delimited import ID from state. The form is
// `tableName,hashKeyValue[,rangeKeyValue]`. Used by the identity-block import
// path to derive a legacy CLI-form import ID from identity attributes.
func (tableItemImportID) Create(d *schema.ResourceData) string {
	parts := []string{
		d.Get(names.AttrTableName).(string),
		d.Get("hash_key_value").(string),
	}
	if v, ok := d.GetOk("range_key_value"); ok {
		parts = append(parts, v.(string))
	}
	return strings.Join(parts, flex.ResourceIdSeparator)
}

// Parse splits a comma-delimited import ID of the form
// `tableName,hashKeyValue[,rangeKeyValue]` into resource attributes. Returns
// the import ID unchanged as the resource ID, which is also the canonical
// internal form. Read recovers the hash and range key names via
// DescribeTable.
//
// Hash key or range key values containing commas must be imported via an
// `import` block with the `identity` attribute set rather than the legacy
// `terraform import` CLI form.
func (tableItemImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, flex.ResourceIdSeparator)
	if len(parts) < 2 || len(parts) > 3 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf(
			"id %q should be of the form tableName%shashKeyValue[%srangeKeyValue]",
			id, flex.ResourceIdSeparator, flex.ResourceIdSeparator,
		)
	}

	result := map[string]any{
		names.AttrTableName: parts[0],
		"hash_key_value":    parts[1],
	}
	if len(parts) == 3 && parts[2] != "" {
		result["range_key_value"] = parts[2]
	}

	return id, result, nil
}
