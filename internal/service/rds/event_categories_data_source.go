// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_event_categories", name="Event Categories")
func dataSourceEventCategories() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEventCategoriesRead,

		Schema: map[string]*schema.Schema{
			"event_categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSourceType: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.SourceType](),
			},
		},
	}
}

func dataSourceEventCategoriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeEventCategoriesInput{}

	if v, ok := d.GetOk(names.AttrSourceType); ok {
		input.SourceType = aws.String(v.(string))
	}

	output, err := findEventCategoriesMaps(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Event Categories: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("event_categories", slices.Concat(tfslices.ApplyToAll(output, func(v types.EventCategoriesMap) []string {
		return v.EventCategories
	})...))

	return diags
}

func findEventCategoriesMaps(ctx context.Context, conn *rds.Client, input *rds.DescribeEventCategoriesInput) ([]types.EventCategoriesMap, error) {
	output, err := conn.DescribeEventCategories(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EventCategoriesMapList, nil
}
