// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_certificate", name="Certificate")
// @Tags(identifierAttribute="arn")
func resourceCertificate() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificate: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			names.AttrCertificateChain: {
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
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"inactive_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrivateKey: {
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CertificateUsageType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := &transfer.ImportCertificateInput{
		Certificate: aws.String(d.Get(names.AttrCertificate).(string)),
		Tags:        getTagsIn(ctx),
		Usage:       awstypes.CertificateUsageType(d.Get("usage").(string)),
	}

	if v, ok := d.GetOk(names.AttrCertificateChain); ok {
		input.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPrivateKey); ok {
		input.PrivateKey = aws.String(v.(string))
	}

	output, err := conn.ImportCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing Transfer Certificate: %s", err)
	}

	d.SetId(aws.ToString(output.CertificateId))

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Certificate (%s): %s", d.Id(), err)
	}

	d.Set("active_date", aws.ToTime(output.ActiveDate).Format(time.RFC3339))
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set(names.AttrCertificateChain, output.CertificateChain)
	d.Set("certificate_id", output.CertificateId)
	d.Set(names.AttrDescription, output.Description)
	d.Set("inactive_date", aws.ToTime(output.InactiveDate).Format(time.RFC3339))
	d.Set("usage", output.Usage)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &transfer.UpdateCertificateInput{
			CertificateId: aws.String(d.Id()),
			Description:   aws.String(d.Get(names.AttrDescription).(string)),
		}

		_, err := conn.UpdateCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Certificate (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	log.Printf("[DEBUG] Deleting Transfer Certificate: %s", d.Id())
	_, err := conn.DeleteCertificate(ctx, &transfer.DeleteCertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCertificateByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedCertificate, error) {
	input := &transfer.DescribeCertificateInput{
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

	if output == nil || output.Certificate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Certificate, nil
}
