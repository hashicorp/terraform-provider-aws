// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudformation_export", name="Export")
func dataSourceExport() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExportRead,

		Schema: map[string]*schema.Schema{
			"exporting_stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	var value *string
	name := d.Get(names.AttrName).(string)
	input := &cloudformation.ListExportsInput{}

	pages := cloudformation.NewListExportsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFormation Exports: %s", err)
		}

		for _, v := range page.Exports {
			if name == aws.ToString(v.Name) {
				d.Set("exporting_stack_id", v.ExportingStackId)
				value = v.Value
				d.Set(names.AttrValue, value)
			}
		}
	}

	if value == nil {
		return sdkdiag.AppendFromErr(diags, tfresource.NewEmptyResultError(name))
	}

	d.SetId(fmt.Sprintf("cloudformation-exports-%s-%s", meta.(*conns.AWSClient).Region, name))

	return diags
}
