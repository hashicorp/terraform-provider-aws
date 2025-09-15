// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrDomain: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"domain", names.AttrID},
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"domain", names.AttrID},
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

	var identifier string
	var tenant *awstypes.DistributionTenant
	var etag *string

	if id, ok := d.GetOk(names.AttrID); ok {
		identifier = id.(string)
		output, err := findDistributionTenantByID(ctx, conn, identifier)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution Tenant by ID (%s): %s", identifier, err)
		}
		tenant = output.DistributionTenant
		etag = output.ETag
	} else {
		identifier = d.Get(names.AttrDomain).(string)
		output, err := findDistributionTenantByDomain(ctx, conn, identifier)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution Tenant by Domain (%s): %s", identifier, err)
		}
		tenant = output.DistributionTenant
		etag = output.ETag
	}

	d.SetId(aws.ToString(tenant.Id))
	d.Set(names.AttrARN, tenant.Arn)
	d.Set("connection_group_id", tenant.ConnectionGroupId)
	d.Set("distribution_id", tenant.DistributionId)
	d.Set("domains", flattenDomains(tenant.Domains))
	d.Set(names.AttrEnabled, tenant.Enabled)
	d.Set("etag", etag)
	d.Set(names.AttrName, tenant.Name)
	d.Set(names.AttrStatus, tenant.Status)
	d.Set("last_modified_time", tenant.LastModifiedTime.String())

	return diags
}

func findDistributionTenantByDomain(ctx context.Context, conn *cloudfront.Client, domain string) (*cloudfront.GetDistributionTenantByDomainOutput, error) {
	input := cloudfront.GetDistributionTenantByDomainInput{
		Domain: aws.String(domain),
	}

	output, err := conn.GetDistributionTenantByDomain(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionTenant == nil || output.DistributionTenant.Domains == nil || output.DistributionTenant.DistributionId == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
