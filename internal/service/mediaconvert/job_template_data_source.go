// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_media_convert_job_template", name="JobTemplate")
// @Tags(identifierAttribute="arn")
func dataSourceJobTemplate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceJobTemplateRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	id := d.Get("id").(string)
	job_template, err := findJobTemplateByName(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert JobTemplate (%s): %s", id, err)
	}

	name := aws.ToString(job_template.Name)
	d.SetId(name)
	d.Set("arn", job_template.Arn)
	d.Set("name", name)

	return diags
}
