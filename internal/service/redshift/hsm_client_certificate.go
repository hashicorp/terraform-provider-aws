// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_hsm_client_certificate", name="HSM Client Certificate")
// @Tags(identifierAttribute="arn")
func resourceHSMClientCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHSMClientCertificateCreate,
		ReadWithoutTimeout:   resourceHSMClientCertificateRead,
		UpdateWithoutTimeout: resourceHSMClientCertificateUpdate,
		DeleteWithoutTimeout: resourceHSMClientCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_client_certificate_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_client_certificate_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHSMClientCertificateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	certIdentifier := d.Get("hsm_client_certificate_identifier").(string)
	input := redshift.CreateHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(certIdentifier),
		Tags:                           getTagsIn(ctx),
	}

	output, err := conn.CreateHsmClientCertificate(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift HSM Client Certificate (%s): %s", certIdentifier, err)
	}

	d.SetId(aws.ToString(output.HsmClientCertificate.HsmClientCertificateIdentifier))

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	out, err := findHSMClientCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift HSM Client Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift HSM Client Certificate (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, hsmClientCertificateARN(ctx, c, d.Id()))
	d.Set("hsm_client_certificate_identifier", out.HsmClientCertificateIdentifier)
	d.Set("hsm_client_certificate_public_key", out.HsmClientCertificatePublicKey)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceHSMClientCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift HSM Client Certificate: %s", d.Id())
	input := redshift.DeleteHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteHsmClientCertificate(ctx, &input)

	if errs.IsA[*awstypes.HsmClientCertificateNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift HSM Client Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func hsmClientCertificateARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "hsmclientcertificate:"+id)
}
