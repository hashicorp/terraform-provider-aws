// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_certificate", name="Certificate")
func dataSourceCertificate() *schema.Resource {
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
			"default_for_new_launches": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"latest_valid_till"},
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"latest_valid_till": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"default_for_new_launches"},
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

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeCertificatesInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.CertificateIdentifier = aws.String(v.(string))
	}

	var certificates []types.Certificate
	var hasDefault bool
	var defaultCertificate string
	pages := rds.NewDescribeCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS Certificates: %s", err)
		}

		if page.DefaultCertificateForNewLaunches != nil && aws.ToString(page.DefaultCertificateForNewLaunches) != "" && !hasDefault {
			hasDefault = true
			defaultCertificate = aws.ToString(page.DefaultCertificateForNewLaunches)
		}

		certificates = append(certificates, page.Certificates...)
	}

	if len(certificates) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS Certificates found")
	}

	// client side filtering
	var certificate *types.Certificate

	if d.Get("latest_valid_till").(bool) {
		slices.SortFunc(certificates, func(a, b types.Certificate) int {
			return a.ValidTill.Compare(*b.ValidTill)
		})
		certificate = &certificates[len(certificates)-1]
	} else if d.Get("default_for_new_launches").(bool) {
		i := slices.IndexFunc(certificates, func(c types.Certificate) bool {
			return aws.ToString(c.CertificateIdentifier) == defaultCertificate
		})

		if i != -1 {
			certificate = &certificates[i]
		} else {
			return sdkdiag.AppendErrorf(diags, "no default RDS Certificate found")
		}
	} else {
		if len(certificates) > 1 {
			return sdkdiag.AppendErrorf(diags, "multiple RDS Certificates match the criteria; try changing search query")
		}
		if len(certificates) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Certificates match the criteria")
		}
		certificate = &certificates[0]
	}

	d.SetId(aws.ToString(certificate.CertificateIdentifier))
	d.Set(names.AttrARN, certificate.CertificateArn)
	d.Set("certificate_type", certificate.CertificateType)
	d.Set("customer_override", certificate.CustomerOverride)
	if certificate.CustomerOverrideValidTill != nil {
		d.Set("customer_override_valid_till", aws.ToTime(certificate.CustomerOverrideValidTill).Format(time.RFC3339))
	}
	d.Set("thumbprint", certificate.Thumbprint)
	if certificate.ValidFrom != nil {
		d.Set("valid_from", aws.ToTime(certificate.ValidFrom).Format(time.RFC3339))
	}
	if certificate.ValidTill != nil {
		d.Set("valid_till", aws.ToTime(certificate.ValidTill).Format(time.RFC3339))
	}

	return diags
}
