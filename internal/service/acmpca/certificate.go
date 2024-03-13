// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/crypto/cryptobyte"
	cryptobyte_asn1 "golang.org/x/crypto/cryptobyte/asn1"
)

// @SDKResource("aws_acmpca_certificate")
func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		DeleteWithoutTimeout: resourceCertificateRevoke,

		// Expects ACM PCA ARN format, e.g:
		// arn:aws:acm-pca:eu-west-1:555885746124:certificate-authority/08322ede-92f9-4200-8f21-c7d12b2b6edb/certificate/a4e9c2aa2ccfab625b1b9136464cd3a6
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				re := regexache.MustCompile(`arn:.+:certificate-authority/[^/]+`)
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SigningAlgorithm](),
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
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ValidityPeriodType](),
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
			"api_passthrough": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const certificateIssueTimeout = 5 * time.Minute
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	certificateAuthorityARN := d.Get("certificate_authority_arn").(string)
	input := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
		Csr:                     []byte(d.Get("certificate_signing_request").(string)),
		IdempotencyToken:        aws.String(id.UniqueId()),
		SigningAlgorithm:        awstypes.SigningAlgorithm(d.Get("signing_algorithm").(string)),
	}
	validity, err := expandValidity(d.Get("validity").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "issuing ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}
	input.Validity = validity

	if v, ok := d.Get("template_arn").(string); ok && v != "" {
		input.TemplateArn = aws.String(v)
	}

	if v, ok := d.Get("api_passthrough").(string); ok && v != "" {
		ap := &awstypes.ApiPassthrough{}
		if err := json.Unmarshal([]byte(v), ap); err != nil {
			return sdkdiag.AppendErrorf(diags, "decoding api_passthrough: %s", err)
		}
		input.ApiPassthrough = ap
	}

	var output *acmpca.IssueCertificateOutput
	err = retry.RetryContext(ctx, certificateAuthorityActiveTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.IssueCertificate(ctx, input)
		if errs.IsAErrorMessageContains[*awstypes.InvalidStateException](err, "The certificate authority is not in a valid state for issuing certificates") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = conn.IssueCertificate(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "issuing ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	d.SetId(aws.ToString(output.CertificateArn))

	// Wait for certificate status to become ISSUED.
	waiter := acmpca.NewCertificateIssuedWaiter(conn)
	params := &acmpca.GetCertificateInput{
		CertificateAuthorityArn: output.CertificateArn,
		CertificateArn:          aws.String(d.Get("certificate_authority_arn").(string)),
	}

	err = waiter.Wait(ctx, params, certificateIssueTimeout)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM PCA Certificate Authority (%s) to issue Certificate (%s), error: %s", certificateAuthorityARN, d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %+v", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(ctx, getCertificateInput)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	block, _ := pem.Decode([]byte(d.Get("certificate").(string)))
	if block == nil {
		log.Printf("[WARN] Failed to parse ACM PCA Certificate (%s)", d.Id())
		return diags
	}

	// Certificate can contain invalid extension values that will prevent full certificate parsing hence revocation
	// but still have serial number that we need in order to revoke it.
	serial, err := getCertificateSerial(block.Bytes)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting ACM PCA Certificate (%s) serial number: %s", d.Id(), err)
	}

	input := &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", serial)),
		RevocationReason:        awstypes.RevocationReasonUnspecified,
	}
	_, err = conn.RevokeCertificate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsA[*awstypes.RequestAlreadyProcessedException](err) ||
		errs.IsA[*awstypes.RequestInProgressException](err) ||
		errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Self-signed certificate can not be revoked") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "revoking ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

// We parse certificate until we get serial number if possible.
// This is partial copy of crypto/x509 package private function parseCertificate
// https://github.com/golang/go/blob/6a70292d1cb3464e5b2c2c03341e5148730a1889/src/crypto/x509/parser.go#L800-L842
func getCertificateSerial(der []byte) (*big.Int, error) {
	malformedCertificateErr := errors.New("malformed certificate")
	input := cryptobyte.String(der)
	if !input.ReadASN1Element(&input, cryptobyte_asn1.SEQUENCE) {
		return nil, malformedCertificateErr
	}
	if !input.ReadASN1(&input, cryptobyte_asn1.SEQUENCE) {
		return nil, malformedCertificateErr
	}

	var tbs cryptobyte.String
	if !input.ReadASN1Element(&tbs, cryptobyte_asn1.SEQUENCE) {
		return nil, malformedCertificateErr
	}
	if !tbs.ReadASN1(&tbs, cryptobyte_asn1.SEQUENCE) {
		return nil, malformedCertificateErr
	}

	var version int
	if !tbs.ReadOptionalASN1Integer(&version, cryptobyte_asn1.Tag(0).Constructed().ContextSpecific(), 0) {
		return nil, errors.New("malformed certificate version")
	}

	serial := new(big.Int)
	if !tbs.ReadASN1Integer(serial) {
		return nil, errors.New("malformed certificate serial number")
	}

	return serial, nil
}

func ValidTemplateARN(v interface{}, k string) (ws []string, errors []error) {
	wsARN, errorsARN := verify.ValidARN(v, k)
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)

	if len(errors) == 0 {
		value := v.(string)
		parsedARN, _ := arn.Parse(value)

		if parsedARN.Service != names.ACMPCAEndpointID {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: service must be \""+names.ACMPCAEndpointID+"\", was %q)", k, value, parsedARN.Service))
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

func expandValidity(l []interface{}) (*awstypes.Validity, error) {
	if len(l) == 0 {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	valueType := m["type"].(string)
	result := &awstypes.Validity{
		Type: awstypes.ValidityPeriodType(valueType),
	}

	i, err := ExpandValidityValue(valueType, m["value"].(string))
	if err != nil {
		return nil, fmt.Errorf("parsing value %q: %w", m["value"].(string), err)
	}
	result.Value = aws.Int64(i)

	return result, nil
}

func ExpandValidityValue(valueType, v string) (int64, error) {
	if valueType == string(awstypes.ValidityPeriodTypeEndDate) {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		v = date.UTC().Format("20060102150405") // YYYYMMDDHHMMSS
	}

	return strconv.ParseInt(v, 10, 64)
}
