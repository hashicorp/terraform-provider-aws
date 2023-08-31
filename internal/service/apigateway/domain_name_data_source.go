// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_api_gateway_domain_name")
func DataSourceDomainName() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainNameRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_arn": {
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
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
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainName := d.Get("domain_name").(string)
	output, err := FindDomainName(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name (%s): %s", domainName, err)
	}

	d.SetId(aws.StringValue(output.DomainName))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("certificate_arn", output.CertificateArn)
	d.Set("certificate_name", output.CertificateName)
	if output.CertificateUploadDate != nil {
		d.Set("certificate_upload_date", output.CertificateUploadDate.Format(time.RFC3339))
	}
	d.Set("cloudfront_domain_name", output.DistributionDomainName)
	d.Set("cloudfront_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID())
	d.Set("domain_name", output.DomainName)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(output.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set("regional_certificate_arn", output.RegionalCertificateArn)
	d.Set("regional_certificate_name", output.RegionalCertificateName)
	d.Set("regional_domain_name", output.RegionalDomainName)
	d.Set("regional_zone_id", output.RegionalHostedZoneId)
	d.Set("security_policy", output.SecurityPolicy)

	if err := d.Set("tags", KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
