// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceSigningCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	createOpts := &iam.UploadSigningCertificateInput{
		CertificateBody: aws.String(d.Get("certificate_body").(string)),
		UserName:        aws.String(d.Get(names.AttrUserName).(string)),
	}

	resp, err := conn.UploadSigningCertificate(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading IAM Signing Certificate: %s", err)
	}

	cert := resp.Certificate
	certId := cert.CertificateId
	d.SetId(fmt.Sprintf("%s:%s", aws.ToString(certId), aws.ToString(cert.UserName)))

	if v, ok := d.GetOk(names.AttrStatus); ok && v.(string) != string(awstypes.StatusTypeActive) {
		updateInput := &iam.UpdateSigningCertificateInput{
			CertificateId: certId,
			UserName:      aws.String(d.Get(names.AttrUserName).(string)),
			Status:        awstypes.StatusType(v.(string)),
		}

		_, err := conn.UpdateSigningCertificate(ctx, updateInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "settings IAM Signing Certificate status: %s", err)
		}
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindSigningCertificate(ctx, conn, userName, certId)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Signing Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	resp := outputRaw.(*awstypes.SigningCertificate)

	d.Set("certificate_body", resp.CertificateBody)
	d.Set("certificate_id", resp.CertificateId)
	d.Set(names.AttrUserName, resp.UserName)
	d.Set(names.AttrStatus, resp.Status)

	return diags
}

func resourceSigningCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	updateInput := &iam.UpdateSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
		Status:        awstypes.StatusType(d.Get(names.AttrStatus).(string)),
	}

	_, err = conn.UpdateSigningCertificate(ctx, updateInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)
	log.Printf("[INFO] Deleting IAM Signing Certificate: %s", d.Id())

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	input := &iam.DeleteSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
	}

	if _, err := conn.DeleteSigningCertificate(ctx, input); err != nil {
		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeSigningCertificateId(id string) (string, string, error) {
	creds := strings.Split(id, ":")
	if len(creds) != 2 {
		return "", "", fmt.Errorf("unknown IAM Signing Certificate ID format")
	}

	certId := creds[0]
	userName := creds[1]

	return certId, userName, nil
}
