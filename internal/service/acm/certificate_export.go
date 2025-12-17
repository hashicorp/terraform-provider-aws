// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_acm_certificate_export", name="Certificate Export")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/acm;acm.ExportCertificateOutput")
func resourceCertificateExport() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateExportCreate,
		ReadWithoutTimeout:   resourceCertificateExportRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"passphrase": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			names.AttrCertificate: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrCertificateChain: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCertificateExportCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	arn := d.Get(names.AttrCertificateARN).(string)
	passphrase := d.Get("passphrase").(string)

	input := &acm.ExportCertificateInput{
		CertificateArn: aws.String(arn),
		Passphrase:     []byte(passphrase),
	}

	output, err := conn.ExportCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "exporting ACM Certificate (%s): %s", arn, err)
	}

	// Use a hash of the certificate ARN and passphrase as the resource ID
	// This ensures a unique ID while being reproducible for the same inputs
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", arn, passphrase)))
	d.SetId(base64.URLEncoding.EncodeToString(hash[:]))

	d.Set(names.AttrCertificate, aws.ToString(output.Certificate))
	d.Set(names.AttrCertificateChain, aws.ToString(output.CertificateChain))
	d.Set(names.AttrPrivateKey, aws.ToString(output.PrivateKey))

	return diags
}

func resourceCertificateExportRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	arn := d.Get(names.AttrCertificateARN).(string)

	// Verify the certificate still exists
	_, err := findCertificateByARN(ctx, conn, arn)

	if err != nil {
		log.Printf("[WARN] ACM Certificate %s not found, removing from state", arn)
		d.SetId("")
		return diags
	}

	// Note: We cannot re-export the certificate on read without the passphrase stored in state
	// The sensitive values are already stored in state from the create operation
	// We only verify the certificate still exists

	return diags
}
