// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_event_categories")
func DataSourceEventCategories() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEventCategoriesRead,

		Schema: map[string]*schema.Schema{
			"event_categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSourceType: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(rds.SourceType_Values(), false),
			},
		},
	}
}

func dataSourceEventCategoriesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeEventCategoriesInput{}

	if v, ok := d.GetOk(names.AttrSourceType); ok {
		input.SourceType = aws.String(v.(string))
	}

	output, err := findEventCategoriesMaps(ctx, conn, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Event Categories: %s", err)
	}

	var eventCategories []string

	for _, v := range output {
		eventCategories = append(eventCategories, aws.StringValueSlice(v.EventCategories)...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("event_categories", eventCategories)

	return diags
}

func findEventCategoriesMaps(ctx context.Context, conn *rds.RDS, input *rds.DescribeEventCategoriesInput) ([]*rds.EventCategoriesMap, error) {
	var output []*rds.EventCategoriesMap

	page, err := conn.DescribeEventCategoriesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, v := range page.EventCategoriesMapList {
		if v != nil {
			output = append(output, v)
		}
	}

	return output, nil
}
