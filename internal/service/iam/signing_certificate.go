package iam

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports

	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSigningCertificate() *schema.Resource {
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
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iam.StatusTypeActive,
				ValidateFunc: validation.StringInSlice(iam.StatusType_Values(), false),
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSigningCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	createOpts := &iam.UploadSigningCertificateInput{
		CertificateBody: aws.String(d.Get("certificate_body").(string)),
		UserName:        aws.String(d.Get("user_name").(string)),
	}

	log.Printf("[DEBUG] Creating IAM Signing Certificate with opts: %s", createOpts)
	resp, err := conn.UploadSigningCertificateWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading IAM Signing Certificate: %s", err)
	}

	cert := resp.Certificate
	certId := cert.CertificateId
	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(certId), aws.StringValue(cert.UserName)))

	if v, ok := d.GetOk("status"); ok && v.(string) != iam.StatusTypeActive {
		updateInput := &iam.UpdateSigningCertificateInput{
			CertificateId: certId,
			UserName:      aws.String(d.Get("user_name").(string)),
			Status:        aws.String(v.(string)),
		}

		_, err := conn.UpdateSigningCertificateWithContext(ctx, updateInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "settings IAM Signing Certificate status: %s", err)
		}
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

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

	resp := outputRaw.(*iam.SigningCertificate)

	d.Set("certificate_body", resp.CertificateBody)
	d.Set("certificate_id", resp.CertificateId)
	d.Set("user_name", resp.UserName)
	d.Set("status", resp.Status)

	return diags
}

func resourceSigningCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	updateInput := &iam.UpdateSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
		Status:        aws.String(d.Get("status").(string)),
	}

	_, err = conn.UpdateSigningCertificateWithContext(ctx, updateInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSigningCertificateRead(ctx, d, meta)...)
}

func resourceSigningCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	log.Printf("[INFO] Deleting IAM Signing Certificate: %s", d.Id())

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Signing Certificate (%s): %s", d.Id(), err)
	}

	input := &iam.DeleteSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
	}

	if _, err := conn.DeleteSigningCertificateWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
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
