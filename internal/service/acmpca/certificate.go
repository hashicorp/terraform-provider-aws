package acmpca

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		DeleteWithoutTimeout: resourceCertificateRevoke,

		// Expects ACM PCA ARN format, e.g:
		// arn:aws:acm-pca:eu-west-1:555885746124:certificate-authority/08322ede-92f9-4200-8f21-c7d12b2b6edb/certificate/a4e9c2aa2ccfab625b1b9136464cd3a6
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				re := regexp.MustCompile(`arn:.+:certificate-authority/[^/]+`)
				authorityARN := re.FindString(d.Id())
				if authorityARN == "" {
					return nil, fmt.Errorf("Unexpected format for ID (%q), expected ACM PCA Certificate ARN", d.Id())
				}

				d.Set("certificate_authority_arn", authorityARN)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"signing_algorithm": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(acmpca.SigningAlgorithm_Values(), false),
			},
			"validity": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(acmpca.ValidityPeriodType_Values(), false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidStringDateOrPositiveInt,
						},
					},
				},
			},
			"template_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: ValidTemplateARN,
			},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()

	certificateAuthorityARN := d.Get("certificate_authority_arn").(string)
	input := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
		Csr:                     []byte(d.Get("certificate_signing_request").(string)),
		IdempotencyToken:        aws.String(resource.UniqueId()),
		SigningAlgorithm:        aws.String(d.Get("signing_algorithm").(string)),
	}
	validity, err := expandValidity(d.Get("validity").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "issuing ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}
	input.Validity = validity

	if v, ok := d.Get("template_arn").(string); ok && v != "" {
		input.TemplateArn = aws.String(v)
	}

	var output *acmpca.IssueCertificateOutput
	err = resource.RetryContext(ctx, certificateAuthorityActiveTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.IssueCertificateWithContext(ctx, input)
		if tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidStateException, "The certificate authority is not in a valid state for issuing certificates") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = conn.IssueCertificateWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "issuing ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	d.SetId(aws.StringValue(output.CertificateArn))

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          output.CertificateArn,
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	err = conn.WaitUntilCertificateIssuedWithContext(ctx, getCertificateInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM PCA Certificate Authority (%s) to issue Certificate (%s), error: %s", certificateAuthorityARN, d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificateWithContext(ctx, getCertificateInput)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM PCA Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	if certificateOutput == nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate (%s): empty response", d.Id())
	}

	d.Set("arn", d.Id())
	d.Set("certificate", certificateOutput.Certificate)
	d.Set("certificate_chain", certificateOutput.CertificateChain)

	return diags
}

func resourceCertificateRevoke(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()

	block, _ := pem.Decode([]byte(d.Get("certificate").(string)))
	if block == nil {
		log.Printf("[WARN] Failed to parse ACM PCA Certificate (%s)", d.Id())
		return diags
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Failed to parse ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	input := &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", cert.SerialNumber)),
		RevocationReason:        aws.String(acmpca.RevocationReasonUnspecified),
	}
	_, err = conn.RevokeCertificateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed certificate can not be revoked") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "revoking ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func ValidTemplateARN(v interface{}, k string) (ws []string, errors []error) {
	wsARN, errorsARN := verify.ValidARN(v, k)
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)

	if len(errors) == 0 {
		value := v.(string)
		parsedARN, _ := arn.Parse(value)

		if parsedARN.Service != acmpca.ServiceName {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: service must be \""+acmpca.ServiceName+"\", was %q)", k, value, parsedARN.Service))
		}

		if parsedARN.Region != "" {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: region must be empty, was %q)", k, value, parsedARN.Region))
		}

		if parsedARN.AccountID != "" {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: account ID must be empty, was %q)", k, value, parsedARN.AccountID))
		}

		if !strings.HasPrefix(parsedARN.Resource, "template/") {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: expected resource to start with \"template/\", was %q)", k, value, parsedARN.Resource))
		}
	}

	return ws, errors
}

func expandValidity(l []interface{}) (*acmpca.Validity, error) {
	if len(l) == 0 {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	valueType := m["type"].(string)
	result := &acmpca.Validity{
		Type: aws.String(valueType),
	}

	i, err := ExpandValidityValue(valueType, m["value"].(string))
	if err != nil {
		return nil, fmt.Errorf("error parsing value %q: %w", m["value"].(string), err)
	}
	result.Value = aws.Int64(i)

	return result, nil
}

func ExpandValidityValue(valueType, v string) (int64, error) {
	if valueType == acmpca.ValidityPeriodTypeEndDate {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		v = date.UTC().Format("20060102150405") // YYYYMMDDHHMMSS
	}

	return strconv.ParseInt(v, 10, 64)
}
