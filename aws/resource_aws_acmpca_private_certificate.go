package aws

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAcmpcaPrivateCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmpcaPrivateCertificateCreate,
		Read:   resourceAwsAcmpcaPrivateCertificateRead,
		Delete: resourceAwsAcmpcaPrivateCertificateRevoke,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
		SchemaVersion: 1,

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
				ValidateFunc: validateArn,
			},
			"certificate_signing_request": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},
			"signing_algorithm": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(acmpca.SigningAlgorithm_Values(), false),
			},
			"validity_length": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"validity_unit": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(acmpca.ValidityPeriodType_Values(), false),
			},
			"template_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAcmPcaTemplateArn,
			},
		},
	}
}

func resourceAwsAcmpcaPrivateCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	input := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		Csr:                     []byte(d.Get("certificate_signing_request").(string)),
		IdempotencyToken:        aws.String(resource.UniqueId()),
		SigningAlgorithm:        aws.String(d.Get("signing_algorithm").(string)),
		Validity: &acmpca.Validity{
			Type:  aws.String(d.Get("validity_unit").(string)),
			Value: aws.Int64(int64(d.Get("validity_length").(int))),
		},
	}
	if v, ok := d.Get("template_arn").(string); ok && v != "" {
		input.TemplateArn = aws.String(v)
	}

	log.Printf("[DEBUG] ACM PCA Issue Certificate: %s", input)
	output, err := conn.IssueCertificate(input)
	if err != nil {
		return fmt.Errorf("error issuing ACM PCA Certificate: %s", err)
	}

	d.SetId(aws.StringValue(output.CertificateArn))

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          output.CertificateArn,
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	err = conn.WaitUntilCertificateIssued(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error waiting for ACM PCA to issue Certificate %q, error: %s", d.Id(), err)
	}

	return resourceAwsAcmpcaPrivateCertificateRead(d, meta)
}

func resourceAwsAcmpcaPrivateCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACM PCA Certificate %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ACM PCA Certificate: %s", err)
	}

	d.Set("arn", d.Id())
	d.Set("certificate", aws.StringValue(certificateOutput.Certificate))
	d.Set("certificate_chain", aws.StringValue(certificateOutput.CertificateChain))

	return nil
}

func resourceAwsAcmpcaPrivateCertificateRevoke(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	block, _ := pem.Decode([]byte(d.Get("certificate").(string)))
	if block == nil {
		log.Printf("[WARN] Failed to parse certificate %q", d.Id())
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("Failed to parse ACM PCA Certificate: %s", err)
	}

	input := &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", cert.SerialNumber)),
		RevocationReason:        aws.String(acmpca.RevocationReasonUnspecified),
	}
	_, err = conn.RevokeCertificate(input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed certificate can not be revoked") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error revoking ACM PCA Certificate: %s", err)
	}

	return nil
}

func validateAcmPcaTemplateArn(v interface{}, k string) (ws []string, errors []error) {
	wsARN, errorsARN := validateArn(v, k)
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
