// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_acmpca_certificate_authority_certificate")
func ResourceCertificateAuthorityCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateAuthorityCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateAuthorityCertificateRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32768),
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate_chain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2097152),
			},
		},
	}
}

func resourceCertificateAuthorityCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn(ctx)

	certificateAuthorityARN := d.Get("certificate_authority_arn").(string)

	input := &acmpca.ImportCertificateAuthorityCertificateInput{
		Certificate:             []byte(d.Get("certificate").(string)),
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}
	if v, ok := d.Get("certificate_chain").(string); ok && v != "" {
		input.CertificateChain = []byte(v)
	}

	_, err := conn.ImportCertificateAuthorityCertificateWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	d.SetId(certificateAuthorityARN)

	return append(diags, resourceCertificateAuthorityCertificateRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn(ctx)

	output, err := FindCertificateAuthorityCertificateByARN(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate Authority Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority Certificate (%s): %s", d.Id(), err)
	}

	d.Set("certificate_authority_arn", d.Id())
	d.Set("certificate", output.Certificate)
	d.Set("certificate_chain", output.CertificateChain)

	return diags
}
