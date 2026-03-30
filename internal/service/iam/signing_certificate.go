// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_signing_certificate", name="Signing Certificate")
func resourceSigningCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningCertificateCreate,
		ReadWithoutTimeout:   resourceSigningCertificateRead,
		UpdateWithoutTimeout: resourceSigningCertificateUpdate,
		DeleteWithoutTimeout: resourceSigningCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNormalizeCertRemoval,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.StatusTypeActive,
				ValidateDiagFunc: enum.Validate[awstypes.StatusType](),
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSigningCertificateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	input := iam.UploadSigningCertificateInput{
		CertificateBody: aws.String(d.Get("certificate_body").(string)),
		UserName:        aws.String(userName),
	}

	output, err := conn.UploadSigningCertificate(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading IAM Signing Certificate: %s", err)
	}

	cert := output.Certificate
	certID := aws.ToString(cert.CertificateId)
	d.SetId(signingCertificateCreateResourceID(certID, userName))

	if v, ok := d.GetOk(names.AttrStatus); ok && v.(string) != string(awstypes.StatusTypeActive) {
		input := iam.UpdateSigningCertificateInput{
			CertificateId: aws.String(certID),
			Status:        awstypes.StatusType(v.(string)),
			UserName:      aws.String(userName),
		}

		_, err := conn.UpdateSigningCertificate(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "settings IAM Signing Certificate (%s) status: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	certID, userName, err := signingCertificateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.SigningCertificate, error) {
		return findSigningCertificateByTwoPartKey(ctx, conn, userName, certID)
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Signing Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	d.Set("certificate_body", output.CertificateBody)
	d.Set("certificate_id", output.CertificateId)
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrUserName, output.UserName)

	return diags
}

func resourceSigningCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	certID, userName, err := signingCertificateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := iam.UpdateSigningCertificateInput{
		CertificateId: aws.String(certID),
		Status:        awstypes.StatusType(d.Get(names.AttrStatus).(string)),
		UserName:      aws.String(userName),
	}
	_, err = conn.UpdateSigningCertificate(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	certID, userName, err := signingCertificateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM Signing Certificate: %s", d.Id())
	input := iam.DeleteSigningCertificateInput{
		CertificateId: aws.String(certID),
		UserName:      aws.String(userName),
	}
	_, err = conn.DeleteSigningCertificate(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

const signingCertificateResourceIDSeparator = ":"

func signingCertificateCreateResourceID(certificateID, userName string) string {
	parts := []string{certificateID, userName}
	id := strings.Join(parts, signingCertificateResourceIDSeparator)

	return id
}

func signingCertificateParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, signingCertificateResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CERTIFICATE-ID%[2]sUSER-NAME", id, signingCertificateResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findSigningCertificateByTwoPartKey(ctx context.Context, conn *iam.Client, userName, certID string) (*awstypes.SigningCertificate, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	return findSigningCertificate(ctx, conn, input, func(v *awstypes.SigningCertificate) bool {
		return aws.ToString(v.CertificateId) == certID
	})
}

func findSigningCertificate(ctx context.Context, conn *iam.Client, input *iam.ListSigningCertificatesInput, filter tfslices.Predicate[*awstypes.SigningCertificate]) (*awstypes.SigningCertificate, error) {
	output, err := findSigningCertificates(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSigningCertificates(ctx context.Context, conn *iam.Client, input *iam.ListSigningCertificatesInput, filter tfslices.Predicate[*awstypes.SigningCertificate]) ([]awstypes.SigningCertificate, error) {
	var output []awstypes.SigningCertificate

	pages := iam.NewListSigningCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Certificates {
			if p := &v; !inttypes.IsZero(p) && filter(p) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
