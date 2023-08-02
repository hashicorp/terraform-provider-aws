// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_rds_certificate")
func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_override": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"customer_override_valid_till": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"latest_valid_till": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"thumbprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_from": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_till": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeCertificatesInput{}

	if v, ok := d.GetOk("id"); ok {
		input.CertificateIdentifier = aws.String(v.(string))
	}

	var certificates []*rds.Certificate

	err := conn.DescribeCertificatesPagesWithContext(ctx, input, func(page *rds.DescribeCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, certificate := range page.Certificates {
			if certificate == nil {
				continue
			}

			certificates = append(certificates, certificate)
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Certificates: %s", err)
	}

	if len(certificates) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS Certificates found")
	}

	// client side filtering
	var certificate *rds.Certificate

	if d.Get("latest_valid_till").(bool) {
		sort.Sort(rdsCertificateValidTillSort(certificates))
		certificate = certificates[len(certificates)-1]
	}

	if len(certificates) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple RDS Certificates match the criteria; try changing search query")
	}

	if certificate == nil && len(certificates) == 1 {
		certificate = certificates[0]
	}

	if certificate == nil {
		return sdkdiag.AppendErrorf(diags, "no RDS Certificates match the criteria")
	}

	d.SetId(aws.StringValue(certificate.CertificateIdentifier))

	d.Set("arn", certificate.CertificateArn)
	d.Set("certificate_type", certificate.CertificateType)
	d.Set("customer_override", certificate.CustomerOverride)

	if certificate.CustomerOverrideValidTill != nil {
		d.Set("customer_override_valid_till", aws.TimeValue(certificate.CustomerOverrideValidTill).Format(time.RFC3339))
	}

	d.Set("thumbprint", certificate.Thumbprint)

	if certificate.ValidFrom != nil {
		d.Set("valid_from", aws.TimeValue(certificate.ValidFrom).Format(time.RFC3339))
	}

	if certificate.ValidTill != nil {
		d.Set("valid_till", aws.TimeValue(certificate.ValidTill).Format(time.RFC3339))
	}

	return diags
}

type rdsCertificateValidTillSort []*rds.Certificate

func (s rdsCertificateValidTillSort) Len() int      { return len(s) }
func (s rdsCertificateValidTillSort) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s rdsCertificateValidTillSort) Less(i, j int) bool {
	if s[i] == nil || s[i].ValidTill == nil {
		return true
	}

	if s[j] == nil || s[j].ValidTill == nil {
		return false
	}

	return (*s[i].ValidTill).Before(*s[j].ValidTill)
}
