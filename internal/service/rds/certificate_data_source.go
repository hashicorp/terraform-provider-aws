// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_certificate")
func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrID: {
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

	if v, ok := d.GetOk(names.AttrID); ok {
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
		slices.SortFunc(certificates, func(a, b *rds.Certificate) int {
			return a.ValidTill.Compare(*b.ValidTill)
		})
		certificate = certificates[len(certificates)-1]
	} else {
		if len(certificates) > 1 {
			return sdkdiag.AppendErrorf(diags, "multiple RDS Certificates match the criteria; try changing search query")
		}
		if len(certificates) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Certificates match the criteria")
		}
		certificate = certificates[0]
	}

	d.SetId(aws.StringValue(certificate.CertificateIdentifier))

	d.Set(names.AttrARN, certificate.CertificateArn)
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
