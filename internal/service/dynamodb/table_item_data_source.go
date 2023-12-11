// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dynamodb_table_item")
func DataSourceTableItem() *schema.Resource {
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
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateTableItem,
			},
			"projection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameTableItem = "Table Item Data Source"
)

func dataSourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	tableName := d.Get("table_name").(string)
	key, err := ExpandTableItemAttributes(d.Get("key").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	id := buildTableItemDataSourceID(tableName, key)
	in := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key:            key,
		TableName:      aws.String(tableName),
	}

	if v, ok := d.GetOk("expression_attribute_names"); ok && len(v.(map[string]interface{})) > 0 {
		in.ExpressionAttributeNames = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("projection_expression"); ok {
		in.ProjectionExpression = aws.String(v.(string))
	}

	out, err := conn.GetItemWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}

	if out.Item == nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}

	d.SetId(id)

	d.Set("expression_attribute_names", aws.StringValueMap(in.ExpressionAttributeNames))
	d.Set("projection_expression", in.ProjectionExpression)
	d.Set("table_name", tableName)

	itemAttrs, err := flattenTableItemAttributes(out.Item)

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}

	d.Set("item", itemAttrs)

	return diags
}

func buildTableItemDataSourceID(tableName string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName}

	for key, element := range attrs {
		id = append(id, key, verify.Base64Encode(element.B))
		id = append(id, aws.StringValue(element.S))
		id = append(id, aws.StringValue(element.N))
	}

	return strings.Join(id, "|")
}
