// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lakeformation_resource")
func DataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.DescribeResourceInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.ResourceArn = aws.String(v.(string))
	}

	output, err := conn.DescribeResource(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.EntityNotFoundException](err) {
		log.Printf("[WARN] Resource Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading data source, Lake Formation Resource (arn: %s): %s", aws.ToString(input.ResourceArn), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return sdkdiag.AppendErrorf(diags, "reading data source, Lake Formation Resource: empty response")
	}

	d.SetId(aws.ToString(input.ResourceArn))
	// d.Set("arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set(names.AttrRoleARN, output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return diags
}
