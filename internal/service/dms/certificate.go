// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_certificate", name="Certificate")
// @Tags(identifierAttribute="certificate_arn")
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
			names.AttrCertificateARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$"), "must start with a letter, only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"certificate_pem": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"certificate_pem", "certificate_wallet"},
			},
			"certificate_wallet": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"certificate_pem", "certificate_wallet"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	certificateID := d.Get("certificate_id").(string)
	input := &dms.ImportCertificateInput{
		CertificateIdentifier: aws.String(certificateID),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("certificate_pem"); ok {
		input.CertificatePem = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_wallet"); ok {
		v, err := itypes.Base64Decode(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.CertificateWallet = v
	}

	_, err := conn.ImportCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Certificate (%s): %s", certificateID, err)
	}

	d.SetId(certificateID)

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	certificate, err := findCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Certificate (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(certificate.CertificateIdentifier))
	d.Set("certificate_id", certificate.CertificateIdentifier)
	d.Set(names.AttrCertificateARN, certificate.CertificateArn)
	if v := aws.ToString(certificate.CertificatePem); v != "" {
		d.Set("certificate_pem", v)
	}
	if certificate.CertificateWallet != nil && len(certificate.CertificateWallet) != 0 {
		d.Set("certificate_wallet", itypes.Base64EncodeOnce(certificate.CertificateWallet))
	}

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	log.Printf("[DEBUG] Deleting DMS Certificate: %s", d.Id())
	_, err := conn.DeleteCertificate(ctx, &dms.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get(names.AttrCertificateARN).(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCertificateByID(ctx context.Context, conn *dms.Client, id string) (*awstypes.Certificate, error) {
	input := &dms.DescribeCertificatesInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []string{id},
			},
		},
	}

	return findCertificate(ctx, conn, input)
}

func findCertificate(ctx context.Context, conn *dms.Client, input *dms.DescribeCertificatesInput) (*awstypes.Certificate, error) {
	output, err := findCertificates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCertificates(ctx context.Context, conn *dms.Client, input *dms.DescribeCertificatesInput) ([]awstypes.Certificate, error) {
	var output []awstypes.Certificate

	pages := dms.NewDescribeCertificatesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Certificates...)
	}

	return output, nil
}
