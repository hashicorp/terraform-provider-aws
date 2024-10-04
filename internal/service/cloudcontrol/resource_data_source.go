// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudcontrol

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudcontrolapi_resource", name="Resource")
func dataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceRead,

		Schema: map[string]*schema.Schema{
			names.AttrIdentifier: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrProperties: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}`), "must be three alphanumeric sections separated by double colons (::)"),
			},
			"type_version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	identifier := d.Get(names.AttrIdentifier).(string)
	typeName := d.Get("type_name").(string)
	resourceDescription, err := findResource(ctx, conn,
		identifier,
		typeName,
		d.Get("type_version_id").(string),
		d.Get(names.AttrRoleARN).(string),
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cloud Control API (%s) Resource (%s): %s", typeName, identifier, err)
	}

	d.SetId(aws.ToString(resourceDescription.Identifier))

	d.Set(names.AttrProperties, resourceDescription.Properties)

	return diags
}
