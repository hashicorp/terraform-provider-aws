// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_certificate", name="Certificate")
// @Tags(identifierAttribute="certificate_id")
func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"active_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"certificate_chain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 2097152),
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"inactive_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
				//ExactlyOneOf: []string{"certificate_chain", "private_key"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"usage": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(transfer.CertificateUsageType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	input := &transfer.ImportCertificateInput{
		Certificate: aws.String(d.Get("certificate").(string)),
		Tags:        getTagsIn(ctx),
		Usage:       aws.String(d.Get("usage").(string)),
	}

	if v, ok := d.GetOk("certificate_chain"); ok {
		input.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_key"); ok {
		input.PrivateKey = aws.String(v.(string))
	}

	output, err := conn.ImportCertificateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing Transfer Certificate: %s", err)
	}

	d.SetId(aws.ToString(output.CertificateId))

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	output, err := FindCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Certificate (%s): %s", d.Id(), err)
	}

	d.Set("active_date", aws.ToTime(output.ActiveDate).Format(time.RFC3339))
	d.Set("certificate", output.Certificate)
	d.Set("certificate_chain", output.CertificateChain)
	d.Set("certificate_id", output.CertificateId)
	d.Set("description", output.Description)
	d.Set("inactive_date", aws.ToTime(output.InactiveDate).Format(time.RFC3339))
	d.Set("usage", output.Usage)
	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.HasChange("description") {
		input := &transfer.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			Description:   aws.String(d.Get("description").(string)),
		}

		_, err := conn.UpdateCertificateWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Certificate (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	log.Printf("[DEBUG] Deleting Transfer Certificate: %s", d.Id())
	_, err := conn.DeleteCertificateWithContext(ctx, &transfer.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Certificate (%s): %s", d.Id(), err)
	}

	return diags
}
