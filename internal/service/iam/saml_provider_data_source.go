// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_iam_saml_provider", name="SAML Provider")
func dataSourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSAMLProviderRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_metadata_document": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	output, err := findSAMLProviderByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM SAML Provider (%s): %s", arn, err)
	}

	name, err := nameFromSAMLProviderARN(arn)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(arn)
	if output.CreateDate != nil {
		d.Set("create_date", aws.ToTime(output.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}
	d.Set("name", name)
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.ToTime(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	tags := KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
