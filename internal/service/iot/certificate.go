// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iot_certificate")
func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,
		Schema: map[string]*schema.Schema{
			"csr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"ca_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	_, okcert := d.GetOk("certificate_pem")
	_, okCA := d.GetOk("ca_pem")

	cert_status := "INACTIVE"
	if d.Get("active").(bool) {
		cert_status = "ACTIVE"
	}

	if _, ok := d.GetOk("csr"); ok {
		log.Printf("[DEBUG] Creating certificate from CSR")
		out, err := conn.CreateCertificateFromCsrWithContext(ctx, &iot.CreateCertificateFromCsrInput{
			CertificateSigningRequest: aws.String(d.Get("csr").(string)),
			SetAsActive:               aws.Bool(d.Get("active").(bool)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating certificate from CSR: %v", err)
		}
		log.Printf("[DEBUG] Created certificate from CSR")

		d.SetId(aws.StringValue(out.CertificateId))
	} else if okcert && okCA {
		log.Printf("[DEBUG] Registering certificate with CA")
		out, err := conn.RegisterCertificateWithContext(ctx, &iot.RegisterCertificateInput{
			CaCertificatePem: aws.String(d.Get("ca_pem").(string)),
			CertificatePem:   aws.String(d.Get("certificate_pem").(string)),
			Status:           aws.String(cert_status),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering certificate with CA: %v", err)
		}
		log.Printf("[DEBUG] Certificate with CA registered")

		d.SetId(aws.StringValue(out.CertificateId))
	} else if okcert {
		log.Printf("[DEBUG] Registering certificate without CA")
		out, err := conn.RegisterCertificateWithoutCAWithContext(ctx, &iot.RegisterCertificateWithoutCAInput{
			CertificatePem: aws.String(d.Get("certificate_pem").(string)),
			Status:         aws.String(cert_status),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering certificate without CA: %v", err)
		}
		log.Printf("[DEBUG] Certificate without CA registered")

		d.SetId(aws.StringValue(out.CertificateId))
	} else {
		log.Printf("[DEBUG] Creating keys and certificate")
		out, err := conn.CreateKeysAndCertificateWithContext(ctx, &iot.CreateKeysAndCertificateInput{
			SetAsActive: aws.Bool(d.Get("active").(bool)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating keys and certificate: %v", err)
		}
		log.Printf("[DEBUG] Created keys and certificate")

		d.SetId(aws.StringValue(out.CertificateId))
		d.Set("public_key", out.KeyPair.PublicKey)
		d.Set("private_key", out.KeyPair.PrivateKey)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	out, err := conn.DescribeCertificateWithContext(ctx, &iot.DescribeCertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading certificate details: %v", err)
	}

	d.Set("active", aws.Bool(aws.StringValue(out.CertificateDescription.Status) == iot.CertificateStatusActive))
	d.Set("arn", out.CertificateDescription.CertificateArn)
	d.Set("certificate_pem", out.CertificateDescription.CertificatePem)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.HasChange("active") {
		status := iot.CertificateStatusInactive
		if d.Get("active").(bool) {
			status = iot.CertificateStatusActive
		}

		_, err := conn.UpdateCertificateWithContext(ctx, &iot.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     aws.String(status),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating certificate: %v", err)
		}
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	_, err := conn.UpdateCertificateWithContext(ctx, &iot.UpdateCertificateInput{
		CertificateId: aws.String(d.Id()),
		NewStatus:     aws.String("INACTIVE"),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "inactivating certificate: %v", err)
	}

	_, err = conn.DeleteCertificateWithContext(ctx, &iot.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting certificate: %v", err)
	}

	return diags
}
