// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_acmpca_certificate", name="Certificate")
func dataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrCertificateChain: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	certificateARN := d.Get(names.AttrARN).(string)
	output, err := findCertificateByTwoPartKey(ctx, conn, certificateARN, d.Get("certificate_authority_arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate (%s): %s", certificateARN, err)
	}

	d.SetId(certificateARN)
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set(names.AttrCertificateChain, output.CertificateChain)

	return diags
}
