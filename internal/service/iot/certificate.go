// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iot_certificate", name="Certificate)
func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"csr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"public_key": {
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

	active := d.Get("active").(bool)
	status := iot.CertificateStatusInactive
	if active {
		status = iot.CertificateStatusActive
	}
	vCert, okCert := d.GetOk("certificate_pem")
	vCA, okCA := d.GetOk("ca_pem")

	if vCSR, okCSR := d.GetOk("csr"); okCSR {
		input := &iot.CreateCertificateFromCsrInput{
			CertificateSigningRequest: aws.String(vCSR.(string)),
			SetAsActive:               aws.Bool(active),
		}

		output, err := conn.CreateCertificateFromCsrWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Certificate from CSR: %s", err)
		}

		d.SetId(aws.StringValue(output.CertificateId))
	} else if okCert && okCA {
		input := &iot.RegisterCertificateInput{
			CaCertificatePem: aws.String(vCA.(string)),
			CertificatePem:   aws.String(vCert.(string)),
			Status:           aws.String(status),
		}

		output, err := conn.RegisterCertificateWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering IoT Certificate with CA: %s", err)
		}

		d.SetId(aws.StringValue(output.CertificateId))
	} else if okCert {
		input := &iot.RegisterCertificateWithoutCAInput{
			CertificatePem: aws.String(vCert.(string)),
			Status:         aws.String(status),
		}

		output, err := conn.RegisterCertificateWithoutCAWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering IoT Certificate without CA: %s", err)
		}

		d.SetId(aws.StringValue(output.CertificateId))
	} else {
		input := &iot.CreateKeysAndCertificateInput{
			SetAsActive: aws.Bool(active),
		}

		output, err := conn.CreateKeysAndCertificateWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Certificate: %s", err)
		}

		d.SetId(aws.StringValue(output.CertificateId))
		d.Set("private_key", output.KeyPair.PrivateKey)
		d.Set("public_key", output.KeyPair.PublicKey)
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
