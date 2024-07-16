// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_api_key", name="API Key")
// @Tags
func dataSourceAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAPIKeyRead,

		Schema: map[string]*schema.Schema{
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrValue: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	id := d.Get(names.AttrID).(string)
	apiKey, err := findAPIKeyByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway API Key (%s): %s", id, err)
	}

	d.SetId(aws.ToString(apiKey.Id))
	d.Set(names.AttrCreatedDate, aws.ToTime(apiKey.CreatedDate).Format(time.RFC3339))
	d.Set("customer_id", apiKey.CustomerId)
	d.Set(names.AttrDescription, apiKey.Description)
	d.Set(names.AttrEnabled, apiKey.Enabled)
	d.Set(names.AttrLastUpdatedDate, aws.ToTime(apiKey.LastUpdatedDate).Format(time.RFC3339))
	d.Set(names.AttrName, apiKey.Name)
	d.Set(names.AttrValue, apiKey.Value)

	setTagsOut(ctx, apiKey.Tags)

	return diags
}
