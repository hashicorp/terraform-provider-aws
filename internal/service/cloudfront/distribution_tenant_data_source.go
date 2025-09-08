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

// @SDKDataSource("aws_cloudfront_distribution_tenant", name="Distribution Tenant")
// @Tags(identifierAttribute="arn")
func dataSourceDistributionTenant() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributionTenantRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domains": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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

func dataSourceDistributionTenantRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	id := d.Get(names.AttrID).(string)
	output, err := findDistributionTenantByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution Tenant (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.DistributionTenant.Id))
	tenant := output.DistributionTenant
	d.Set(names.AttrARN, tenant.Arn)
	d.Set("connection_group_id", tenant.ConnectionGroupId)
	d.Set("distribution_id", tenant.DistributionId)
	d.Set("domains", flattenDomains(tenant.Domains))
	d.Set(names.AttrEnabled, tenant.Enabled)
	d.Set("etag", output.ETag)
	d.Set(names.AttrName, tenant.Name)
	d.Set(names.AttrStatus, tenant.Status)
	d.Set("last_modified_time", tenant.LastModifiedTime.String())

	return diags
}
