// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_certificate", name="Certificate)
func resourceCertificate() *schema.Resource {
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
			names.AttrPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrPublicKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	active := d.Get("active").(bool)
	status := awstypes.CertificateStatusInactive
	if active {
		status = awstypes.CertificateStatusActive
	}
	vCert, okCert := d.GetOk("certificate_pem")
	vCA, okCA := d.GetOk("ca_pem")

	if vCSR, okCSR := d.GetOk("csr"); okCSR {
		input := &iot.CreateCertificateFromCsrInput{
			CertificateSigningRequest: aws.String(vCSR.(string)),
			SetAsActive:               active,
		}

		output, err := conn.CreateCertificateFromCsr(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Certificate from CSR: %s", err)
		}

		d.SetId(aws.ToString(output.CertificateId))
	} else if okCert && okCA {
		input := &iot.RegisterCertificateInput{
			CaCertificatePem: aws.String(vCA.(string)),
			CertificatePem:   aws.String(vCert.(string)),
			Status:           status,
		}

		output, err := conn.RegisterCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering IoT Certificate with CA: %s", err)
		}

		d.SetId(aws.ToString(output.CertificateId))
	} else if okCert {
		input := &iot.RegisterCertificateWithoutCAInput{
			CertificatePem: aws.String(vCert.(string)),
			Status:         status,
		}

		output, err := conn.RegisterCertificateWithoutCA(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "registering IoT Certificate without CA: %s", err)
		}

		d.SetId(aws.ToString(output.CertificateId))
	} else {
		input := &iot.CreateKeysAndCertificateInput{
			SetAsActive: active,
		}

		output, err := conn.CreateKeysAndCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Certificate: %s", err)
		}

		d.SetId(aws.ToString(output.CertificateId))
		d.Set(names.AttrPrivateKey, output.KeyPair.PrivateKey)
		d.Set(names.AttrPublicKey, output.KeyPair.PublicKey)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Certificate (%s): %s", d.Id(), err)
	}

	certificateDescription := output.CertificateDescription
	d.Set("active", certificateDescription.Status == awstypes.CertificateStatusActive)
	d.Set(names.AttrARN, certificateDescription.CertificateArn)
	d.Set("ca_certificate_id", certificateDescription.CaCertificateId)
	d.Set("certificate_pem", certificateDescription.CertificatePem)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	status := awstypes.CertificateStatusInactive
	if d.Get("active").(bool) {
		status = awstypes.CertificateStatusActive
	}
	input := &iot.UpdateCertificateInput{
		CertificateId: aws.String(d.Id()),
		NewStatus:     status,
	}

	_, err := conn.UpdateCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Certificate (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.Get("active").(bool) {
		_, err := conn.UpdateCertificate(ctx, &iot.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     awstypes.CertificateStatusInactive,
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT Certificate (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT Certificate: %s", d.Id())
	_, err := conn.DeleteCertificate(ctx, &iot.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCertificateByID(ctx context.Context, conn *iot.Client, id string) (*iot.DescribeCertificateOutput, error) {
	input := &iot.DescribeCertificateInput{
		CertificateId: aws.String(id),
	}

	output, err := conn.DescribeCertificate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
