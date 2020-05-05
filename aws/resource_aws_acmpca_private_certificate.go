package aws

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					acmpca.SigningAlgorithmSha256withecdsa,
					acmpca.SigningAlgorithmSha256withrsa,
					acmpca.SigningAlgorithmSha384withecdsa,
					acmpca.SigningAlgorithmSha384withrsa,
					acmpca.SigningAlgorithmSha512withecdsa,
					acmpca.SigningAlgorithmSha512withrsa,
				}, false),
			},
			"validity_length": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"validity_unit": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					acmpca.ValidityPeriodTypeAbsolute,
					acmpca.ValidityPeriodTypeDays,
					acmpca.ValidityPeriodTypeEndDate,
					acmpca.ValidityPeriodTypeMonths,
					acmpca.ValidityPeriodTypeYears,
				}, false),
			},
			"template_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"arn:aws:acm-pca:::template/EndEntityCertificate/V1",
					"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen0/V1",
					"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen1/V1",
					"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen2/V1",
					"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen3/V1",
					"arn:aws:acm-pca:::template/RootCACertificate/V1",
				}, false),
				Default: "arn:aws:acm-pca:::template/EndEntityCertificate/V1",
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
		TemplateArn:             aws.String(d.Get("template_arn").(string)),
		Validity: &acmpca.Validity{
			Type:  aws.String(d.Get("validity_unit").(string)),
			Value: aws.Int64(int64(d.Get("validity_length").(int))),
		},
	}

	log.Printf("[DEBUG] ACMPCA Issue Certificate: %s", input)
	var output *acmpca.IssueCertificateOutput
	output, err := conn.IssueCertificate(input)
	if err != nil {
		return fmt.Errorf("error issuing ACMPCA Certificate: %s", err)
	}

	d.SetId(aws.StringValue(output.CertificateArn))

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          output.CertificateArn,
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	err = conn.WaitUntilCertificateIssued(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error waiting for ACMPCA to issue Certificate %q, error: %s", d.Id(), err)
	}

	return resourceAwsAcmpcaPrivateCertificateRead(d, meta)
}

func resourceAwsAcmpcaPrivateCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACMPCA Certificate %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ACMPCA Certificate: %s", err)
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
		return fmt.Errorf("Failed to parse ACMPCA Certificate: %s", err)
	}

	input := &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", cert.SerialNumber)),
		RevocationReason:        aws.String(acmpca.RevocationReasonUnspecified),
	}

	log.Printf("[DEBUG] Revoking ACMPCA Certificate: %s", input)
	_, err = conn.RevokeCertificate(input)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") ||
			isAWSErr(err, acmpca.ErrCodeRequestAlreadyProcessedException, "") ||
			isAWSErr(err, acmpca.ErrCodeRequestInProgressException, "") {
			return nil
		}
		return fmt.Errorf("error revoking ACMPCA Certificate: %s", err)
	}

	return nil
}
