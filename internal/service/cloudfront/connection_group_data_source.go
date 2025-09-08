// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_connection_group", name="Connection Group")
// @Tags(identifierAttribute="arn")
func dataSourceConnectionGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionGroupRead,

		Schema: map[string]*schema.Schema{
			"anycast_ip_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipv6_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"routing_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	id := d.Get(names.AttrID).(string)
	output, err := findConnectionGroupByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Connection Group (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.ConnectionGroup.Id))
	connectionGroup := output.ConnectionGroup
	d.Set("anycast_ip_list_id", connectionGroup.AnycastIpListId)
	d.Set(names.AttrARN, connectionGroup.Arn)
	d.Set(names.AttrEnabled, connectionGroup.Enabled)
	d.Set("etag", output.ETag)
	d.Set("ipv6_enabled", connectionGroup.Ipv6Enabled)
	d.Set("is_default", connectionGroup.IsDefault)
	d.Set("last_modified_time", connectionGroup.LastModifiedTime.String())
	d.Set(names.AttrName, connectionGroup.Name)
	d.Set("routing_endpoint", connectionGroup.RoutingEndpoint)
	d.Set(names.AttrStatus, connectionGroup.Status)

	return diags
}
