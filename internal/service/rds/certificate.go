// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_rds_certificate")
func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Schema: map[string]*schema.Schema{
			"certificate_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds_sdkv2.ModifyCertificatesInput{
		CertificateIdentifier: aws.String(d.Get("certificate_identifier").(string)),
	}

	output, err := conn.ModifyCertificates(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Overriding the system-default SSL/TLS certificate to (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(output.Certificate.CertificateIdentifier))

	return diags
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds_sdkv2.DescribeCertificatesInput{
		CertificateIdentifier: aws.String(d.Get("certificate_identifier").(string)),
	}

	output, err := conn.DescribeCertificates(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Reading the system-default SSL/TLS certificate: %s", err)
	}
	d.Set("certificate_identifier", output.DefaultCertificateForNewLaunches)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds_sdkv2.ModifyCertificatesInput{
		CertificateIdentifier: aws.String(d.Get("certificate_identifier").(string)),
	}

	_, err := conn.ModifyCertificates(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Overriding the system-default SSL/TLS certificate to (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds_sdkv2.ModifyCertificatesInput{
		RemoveCustomerOverride: aws.Bool(true),
	}

	_, err := conn.ModifyCertificates(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Removing the custom override to the system-default SSL/TLS with certificate (%s): %s", d.Id(), err)
	}

	return diags
}
