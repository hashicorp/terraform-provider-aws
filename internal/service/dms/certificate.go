// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_certificate", name="Certificate")
// @Tags(identifierAttribute="certificate_arn")
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
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

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

	_, err := conn.ImportCertificateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Certificate (%s): %s", certificateID, err)
	}

	d.SetId(certificateID)

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	certificate, err := FindCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Certificate (%s): %s", d.Id(), err)
	}

	resourceCertificateSetState(d, certificate)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	log.Printf("[DEBUG] Deleting DMS Certificate: %s", d.Id())
	_, err := conn.DeleteCertificateWithContext(ctx, &dms.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get(names.AttrCertificateARN).(string)),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceCertificateSetState(d *schema.ResourceData, cert *dms.Certificate) {
	d.SetId(aws.StringValue(cert.CertificateIdentifier))

	d.Set("certificate_id", cert.CertificateIdentifier)
	d.Set(names.AttrCertificateARN, cert.CertificateArn)

	if aws.StringValue(cert.CertificatePem) != "" {
		d.Set("certificate_pem", cert.CertificatePem)
	}
	if cert.CertificateWallet != nil && len(cert.CertificateWallet) != 0 {
		d.Set("certificate_wallet", itypes.Base64EncodeOnce(cert.CertificateWallet))
	}
}

func FindCertificateByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.Certificate, error) {
	input := &dms.DescribeCertificatesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []*string{aws.String(id)},
			},
		},
	}

	return findCertificate(ctx, conn, input)
}

func findCertificate(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeCertificatesInput) (*dms.Certificate, error) {
	output, err := findCertificates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findCertificates(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeCertificatesInput) ([]*dms.Certificate, error) {
	var output []*dms.Certificate

	err := conn.DescribeCertificatesPagesWithContext(ctx, input, func(page *dms.DescribeCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Certificates {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
