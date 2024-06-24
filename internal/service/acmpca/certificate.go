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
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
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

// @SDKResource("aws_acmpca_certificate", name="Certificate")
func resourceCertificate() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrCertificateChain: {
				Type:     schema.TypeString,
				Computed: true,
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
				ValidateDiagFunc: enum.Validate[types.SigningAlgorithm](),
			},
			"template_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validTemplateARN,
			},
			"validity": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.ValidityPeriodType](),
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidStringDateOrPositiveInt,
						},
					},
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
	inputI := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
		Csr:                     []byte(d.Get("certificate_signing_request").(string)),
		IdempotencyToken:        aws.String(id.UniqueId()),
		SigningAlgorithm:        types.SigningAlgorithm(d.Get("signing_algorithm").(string)),
	}

	if v, ok := d.Get("api_passthrough").(string); ok && v != "" {
		ap := &types.ApiPassthrough{}
		if err := json.Unmarshal([]byte(v), ap); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		inputI.ApiPassthrough = ap
	}

	if v, ok := d.Get("template_arn").(string); ok && v != "" {
		inputI.TemplateArn = aws.String(v)
	}

	if validity, err := expandValidity(d.Get("validity").([]interface{})); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		inputI.Validity = validity
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidStateException](ctx, certificateAuthorityActiveTimeout, func() (interface{}, error) {
		return conn.IssueCertificate(ctx, inputI)
	}, "The certificate authority is not in a valid state for issuing certificates")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "issuing ACM PCA Certificate with Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	d.SetId(aws.ToString(outputRaw.(*acmpca.IssueCertificateOutput).CertificateArn))

	// Wait for certificate status to become ISSUED.
	inputG := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}
	err = acmpca.NewCertificateIssuedWaiter(conn).Wait(ctx, inputG, certificateIssueTimeout)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM PCA Certificate Authority (%s) to issue Certificate (%s), error: %s", certificateAuthorityARN, d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	output, err := findCertificateByTwoPartKey(ctx, conn, d.Id(), d.Get("certificate_authority_arn").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set("certificate_authority_arn", d.Get("certificate_authority_arn").(string))
	d.Set(names.AttrCertificateChain, output.CertificateChain)

	return diags
}

func resourceCertificateRevoke(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	block, _ := pem.Decode([]byte(d.Get(names.AttrCertificate).(string)))
	if block == nil {
		log.Printf("[WARN] Failed to parse ACM PCA Certificate (%s)", d.Id())
		return diags
	}

	// Certificate can contain invalid extension values that will prevent full certificate parsing hence revocation
	// but still have serial number that we need in order to revoke it.
	serial, err := getCertificateSerial(block.Bytes)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Revoking ACM PCA Certificate: %s", d.Id())
	_, err = conn.RevokeCertificate(ctx, &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", serial)),
		RevocationReason:        types.RevocationReasonUnspecified,
	})

	if errs.IsA[*types.ResourceNotFoundException](err) ||
		errs.IsA[*types.RequestAlreadyProcessedException](err) ||
		errs.IsA[*types.RequestInProgressException](err) ||
		errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "Self-signed certificate can not be revoked") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "revoking ACM PCA Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCertificateByTwoPartKey(ctx context.Context, conn *acmpca.Client, certificateARN, certificateAuthorityARN string) (*acmpca.GetCertificateOutput, error) {
	input := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(certificateARN),
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	return findCertificate(ctx, conn, input)
}

func findCertificate(ctx context.Context, conn *acmpca.Client, input *acmpca.GetCertificateInput) (*acmpca.GetCertificateOutput, error) {
	output, err := conn.GetCertificate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
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

func validTemplateARN(v interface{}, k string) (ws []string, errors []error) {
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

func expandValidity(l []interface{}) (*types.Validity, error) {
	if len(l) == 0 {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	valueType := m[names.AttrType].(string)
	result := &types.Validity{
		Type: types.ValidityPeriodType(valueType),
	}

	i, err := ExpandValidityValue(valueType, m[names.AttrValue].(string))
	if err != nil {
		return nil, fmt.Errorf("parsing value %q: %w", m[names.AttrValue].(string), err)
	}
	result.Value = aws.Int64(i)

	return result, nil
}

func ExpandValidityValue(valueType, v string) (int64, error) {
	if valueType == string(types.ValidityPeriodTypeEndDate) {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		v = date.UTC().Format("20060102150405") // YYYYMMDDHHMMSS
	}

	return strconv.ParseInt(v, 10, 64)
}
