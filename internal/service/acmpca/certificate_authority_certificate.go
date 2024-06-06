// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_acmpca_certificate_authority_certificate", name="Certificate Authority Certificate")
func resourceCertificateAuthorityCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateAuthorityCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateAuthorityCertificateRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCertificate: {
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
			names.AttrCertificateChain: {
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
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	certificateAuthorityARN := d.Get("certificate_authority_arn").(string)
	input := &acmpca.ImportCertificateAuthorityCertificateInput{
		Certificate:             []byte(d.Get(names.AttrCertificate).(string)),
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	if v, ok := d.Get(names.AttrCertificateChain).(string); ok && v != "" {
		input.CertificateChain = []byte(v)
	}

	_, err := conn.ImportCertificateAuthorityCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	d.SetId(certificateAuthorityARN)

	return append(diags, resourceCertificateAuthorityCertificateRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	output, err := findCertificateAuthorityCertificateByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate Authority Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority Certificate (%s): %s", d.Id(), err)
	}

	d.Set("certificate_authority_arn", d.Id())
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set(names.AttrCertificateChain, output.CertificateChain)

	return diags
}

func findCertificateAuthorityCertificateByARN(ctx context.Context, conn *acmpca.Client, arn string) (*acmpca.GetCertificateAuthorityCertificateOutput, error) {
	input := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.GetCertificateAuthorityCertificate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
