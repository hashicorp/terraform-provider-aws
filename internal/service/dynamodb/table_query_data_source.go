// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dynamodb_table_query")
func DataSourceTableQuery() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTableQueryRead,

		Schema: map[string]*schema.Schema{
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_condition_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expression_attribute_values": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"consistent_read": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"expression_attribute_names": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"index_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"output_limit": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"projection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scan_index_forward": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"select": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALL_ATTRIBUTES", "ALL_PROJECTED_ATTRIBUTES", "SPECIFIC_ATTRIBUTES", "COUNT"}, false),
			},
			"items": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"scanned_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"item_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"query_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

type AttributeValue struct {
	B    []byte                     `json:"B,omitempty"`
	BOOL *bool                      `json:"BOOL,omitempty"`
	BS   [][]byte                   `json:"BS,omitempty"`
	L    []*AttributeValue          `json:"L,omitempty"`
	M    map[string]*AttributeValue `json:"M,omitempty"`
	N    *string                    `json:"N,omitempty"`
	NS   []*string                  `json:"NS,omitempty"`
	NULL *bool                      `json:"NULL,omitempty"`
	S    *string                    `json:"S,omitempty"`
	SS   []*string                  `json:"SS,omitempty"`
}

func ConvertJSONToAttributeValue(jsonStr string) (*AttributeValue, error) {
	data := AttributeValue{}

	unescapedJSONStr, err := strconv.Unquote(jsonStr)
	if err != nil {
		// If unquoting fails, assume the string is unescaped and use it as-is
		unescapedJSONStr = jsonStr
	}

	err = json.Unmarshal([]byte(unescapedJSONStr), &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return &data, nil
}

func ConvertToDynamoAttributeValue(av *AttributeValue) (*dynamodb.AttributeValue, error) {
	if av == nil {
		return nil, nil
	}
	dynamoAV := &dynamodb.AttributeValue{}
	if av.B != nil {
		dynamoAV.B = av.B
	}

	if av.BOOL != nil {
		dynamoAV.BOOL = av.BOOL
	}

	if av.BS != nil {
		var bs [][]byte
		for _, item := range av.BS {
			bs = append(bs, item)
		}
		dynamoAV.BS = bs
	}

	if av.L != nil {
		var l []*dynamodb.AttributeValue
		for _, item := range av.L {
			dynamoItem, err := ConvertToDynamoAttributeValue(item)
			if err != nil {
				return nil, err
			}
			l = append(l, dynamoItem)
		}
		dynamoAV.L = l
	}

	if av.M != nil {
		m := make(map[string]*dynamodb.AttributeValue)
		for k, v := range av.M {
			dynamoItem, err := ConvertToDynamoAttributeValue(v)
			if err != nil {
				return nil, err
			}
			m[k] = dynamoItem
		}
		dynamoAV.M = m
	}

	if av.N != nil {
		dynamoAV.N = av.N
	}

	if av.NS != nil {
		var ns []*string
		for _, item := range av.NS {
			ns = append(ns, item)
		}
		dynamoAV.NS = ns
	}

	if av.NULL != nil {
		dynamoAV.NULL = av.NULL
	}

	if av.S != nil {
		dynamoAV.S = av.S
	}

	if av.SS != nil {
		var ss []*string
		for _, item := range av.SS {
			ss = append(ss, item)
		}
		dynamoAV.SS = ss
	}

	return dynamoAV, nil
}

func dataSourceTableQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	tableName := d.Get("table_name").(string)
	keyConditionExpression := d.Get("key_condition_expression").(string)

	in := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String(keyConditionExpression),
	}

	if v, ok := d.GetOkExists("consistent_read"); ok {
		in.ConsistentRead = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOkExists("scan_index_forward"); ok {
		in.ScanIndexForward = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("filter_expression"); ok {
		in.FilterExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("index_name"); ok {
		in.IndexName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("projection_expression"); ok {
		in.ProjectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("select"); ok {
		in.Select = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expression_attribute_names"); ok && len(v.(map[string]interface{})) > 0 {
		in.ExpressionAttributeNames = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("expression_attribute_values"); ok && len(v.(map[string]interface{})) > 0 {
		expressionAttributeValues := flex.ExpandStringMap(v.(map[string]interface{}))
		attributeValues := make(map[string]*dynamodb.AttributeValue)
		for key, value := range expressionAttributeValues {
			jsonData, err := json.Marshal(value)
			if err != nil {
				return diag.FromErr(err)
			}
			attributeValue, err := ConvertJSONToAttributeValue(string(jsonData))
			if err != nil {
				return diag.FromErr(err)
			}
			dynamoAttributeValue, err := ConvertToDynamoAttributeValue(attributeValue)
			if err != nil {
				return diag.FromErr(err)
			}
			attributeValues[key] = dynamoAttributeValue
		}
		in.ExpressionAttributeValues = attributeValues
	}

	var outputLimit *int
	if v, ok := d.GetOk("output_limit"); ok {
		value := v.(int)
		outputLimit = &value
	}

	id := buildTableQueryDataSourceID(tableName, keyConditionExpression)
	d.SetId(id)

	var flattenedItems []string
	itemsProcessed := int64(0)
	scannedCount := int64(0)
	queryCount := int64(0)
	itemCount := int64(0)

	for {
		out, err := conn.QueryWithContext(ctx, in)
		if err != nil {
			return diag.FromErr(err)
		}

		queryCount += 1
		scannedCount += aws.Int64Value(out.ScannedCount)
		itemCount += aws.Int64Value(out.Count)
		for _, item := range out.Items {
			flattened, err := flattenTableItemAttributes(item)
			if err != nil {
				return create.DiagError(names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
			}
			flattenedItems = append(flattenedItems, flattened)

			itemsProcessed++
			if (outputLimit != nil) && (itemsProcessed >= int64(*outputLimit)) {
				itemCount = int64(*outputLimit)
				goto ExitLoop
			}
		}
		in.ExclusiveStartKey = out.LastEvaluatedKey

		if out.LastEvaluatedKey == nil || len(out.LastEvaluatedKey) == 0 {
			break
		}
	}
ExitLoop:
	d.Set("items", flattenedItems)
	d.Set("item_count", itemCount)
	d.Set("query_count", queryCount)
	d.Set("scanned_count", scannedCount)
	return nil
}

func buildTableQueryDataSourceID(tableName, keyConditionExpression string) string {
	id := []string{tableName}
	if keyConditionExpression != "" {
		id = append(id, "KeyConditionExpression", keyConditionExpression)
	}
	id = append(id, fmt.Sprintf("%d", time.Now().UnixNano()))
	return strings.Join(id, "|")
}
