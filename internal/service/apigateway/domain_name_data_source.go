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

// @SDKDataSource("aws_api_gateway_domain_name", name="Domain Name")
// @Tags
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain()")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
// @Testing(tagsIdentifierAttribute="arn")
func dataSourceDomainName() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainNameRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_name_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_certificate_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	var domainNameID string
	domainName := d.Get(names.AttrDomainName).(string)
	if v, ok := d.GetOk("domain_name_id"); ok {
		domainNameID = v.(string)
	}
	output, err := findDomainNameByTwoPartKey(ctx, conn, domainName, domainNameID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name (%s): %s", domainNameCreateResourceID(domainName, domainNameID), err)
	}

	d.SetId(aws.ToString(output.DomainName))
	d.Set(names.AttrARN, domainNameARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrCertificateARN, output.CertificateArn)
	d.Set("certificate_name", output.CertificateName)
	if output.CertificateUploadDate != nil {
		d.Set("certificate_upload_date", output.CertificateUploadDate.Format(time.RFC3339))
	}
	d.Set("cloudfront_domain_name", output.DistributionDomainName)
	d.Set("cloudfront_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set(names.AttrDomainName, output.DomainName)
	d.Set("domain_name_id", output.DomainNameId)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(output.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set(names.AttrPolicy, output.Policy)
	d.Set("regional_certificate_arn", output.RegionalCertificateArn)
	d.Set("regional_certificate_name", output.RegionalCertificateName)
	d.Set("regional_domain_name", output.RegionalDomainName)
	d.Set("regional_zone_id", output.RegionalHostedZoneId)
	d.Set("security_policy", output.SecurityPolicy)

	setTagsOut(ctx, output.Tags)

	return diags
}
