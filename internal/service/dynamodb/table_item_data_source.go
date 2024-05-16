// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dynamodb_table_item", name="Table Item")
func dataSourceTableItem() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTableItemRead,

		Schema: map[string]*schema.Schema{
			"expression_attribute_names": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"item": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateTableItem,
			},
			"projection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrTableName).(string)
	key, err := expandTableItemAttributes(d.Get(names.AttrKey).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	id := createTableItemDataSourceID(tableName, key)
	input := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key:            key,
		TableName:      aws.String(tableName),
	}

	if v, ok := d.GetOk("expression_attribute_names"); ok && len(v.(map[string]interface{})) > 0 {
		input.ExpressionAttributeNames = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("projection_expression"); ok {
		input.ProjectionExpression = aws.String(v.(string))
	}

	item, err := findTableItem(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Item (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("expression_attribute_names", input.ExpressionAttributeNames)
	d.Set("projection_expression", input.ProjectionExpression)
	d.Set(names.AttrTableName, tableName)

	itemAttrs, err := flattenTableItemAttributes(item)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("item", itemAttrs)

	return diags
}

func createTableItemDataSourceID(tableName string, attrs map[string]awstypes.AttributeValue) string {
	id := []string{tableName}

	for k, v := range attrs {
		switch v := v.(type) {
		case *awstypes.AttributeValueMemberB:
			id = append(id, k, itypes.Base64EncodeOnce(v.Value))
		case *awstypes.AttributeValueMemberN:
			id = append(id, v.Value)
		case *awstypes.AttributeValueMemberS:
			id = append(id, v.Value)
		}
	}

	return strings.Join(id, "|")
}
