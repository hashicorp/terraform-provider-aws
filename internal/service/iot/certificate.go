// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificate_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				ForceNew:  true,
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

	output, err := FindCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Certificate (%s): %s", d.Id(), err)
	}

	certificateDescription := output.CertificateDescription
	d.Set("active", aws.StringValue(certificateDescription.Status) == iot.CertificateStatusActive)
	d.Set(names.AttrARN, certificateDescription.CertificateArn)
	d.Set("ca_certificate_id", certificateDescription.CaCertificateId)
	d.Set("certificate_pem", certificateDescription.CertificatePem)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	status := iot.CertificateStatusInactive
	if d.Get("active").(bool) {
		status = iot.CertificateStatusActive
	}
	input := &iot.UpdateCertificateInput{
		CertificateId: aws.String(d.Id()),
		NewStatus:     aws.String(status),
	}

	_, err := conn.UpdateCertificateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Certificate (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.Get("active").(bool) {
		log.Printf("[DEBUG] Disabling IoT Certificate: %s", d.Id())
		_, err := conn.UpdateCertificateWithContext(ctx, &iot.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     aws.String(iot.CertificateStatusInactive),
		})

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT Certificate (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT Certificate: %s", d.Id())
	_, err := conn.DeleteCertificateWithContext(ctx, &iot.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func FindCertificateByID(ctx context.Context, conn *iot.IoT, id string) (*iot.DescribeCertificateOutput, error) {
	input := &iot.DescribeCertificateInput{
		CertificateId: aws.String(id),
	}

	output, err := conn.DescribeCertificateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CertificateDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
