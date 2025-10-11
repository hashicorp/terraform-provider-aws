// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rdsdata_query", name="Query")
func dataSourceQuery() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueryRead,

		Schema: map[string]*schema.Schema{
			names.AttrDatabase: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sql": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
						"type_hint": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"records": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_records_updated": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := rdsdata.NewFromConfig(meta.(*conns.AWSClient).AwsConfig(ctx))

	input := &rdsdata.ExecuteStatementInput{
		ResourceArn:     aws.String(d.Get("resource_arn").(string)),
		SecretArn:       aws.String(d.Get("secret_arn").(string)),
		Sql:             aws.String(d.Get("sql").(string)),
		FormatRecordsAs: types.RecordsFormatTypeJson,
	}

	if v, ok := d.GetOk(names.AttrDatabase); ok {
		input.Database = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok && len(v.([]interface{})) > 0 {
		input.Parameters = expandSqlParameters(v.([]interface{}))
	}

	output, err := conn.ExecuteStatement(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing RDS Data API statement: %s", err)
	}

	d.SetId(aws.ToString(input.ResourceArn) + ":" + aws.ToString(input.Sql))
	d.Set("records", aws.ToString(output.FormattedRecords))
	d.Set("number_of_records_updated", output.NumberOfRecordsUpdated)

	return diags
}

func expandSqlParameters(tfList []interface{}) []types.SqlParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.SqlParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.SqlParameter{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}

		if v, ok := tfMap["type_hint"].(string); ok && v != "" {
			apiObject.TypeHint = types.TypeHint(v)
		}

		// Convert value to Field type
		valueStr := tfMap[names.AttrValue].(string)
		var field types.Field

		// Try to parse as JSON first, otherwise treat as string
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(valueStr), &jsonValue); err == nil {
			switch v := jsonValue.(type) {
			case string:
				field = &types.FieldMemberStringValue{Value: v}
			case float64:
				field = &types.FieldMemberDoubleValue{Value: v}
			case bool:
				field = &types.FieldMemberBooleanValue{Value: v}
			default:
				field = &types.FieldMemberStringValue{Value: valueStr}
			}
		} else {
			field = &types.FieldMemberStringValue{Value: valueStr}
		}

		apiObject.Value = field
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
