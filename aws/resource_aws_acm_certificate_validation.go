package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

func resourceAwsAcmCertificateValidation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmCertificateValidationCreate,
		Read:   resourceAwsAcmCertificateValidationRead,
		Delete: resourceAwsAcmCertificateValidationDelete,

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"validation_record_fqdn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"timeout": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "45m",
			},
		},
	}
}

func resourceAwsAcmCertificateValidationCreate(d *schema.ResourceData, meta interface{}) error {
	// TODO: check that validation_record_fqdn (if set) matches the domain validation information for cert
	// TODO: check that certificate is AMAZON_ISSUED and has Domain Validation enabled (should we also allow to wait for E-Mail validation?)

	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return err
	}

	return resource.Retry(timeout, func() *resource.RetryError {
		acmconn := meta.(*AWSClient).acmconn

		certificate_arn := d.Get("certificate_arn").(string)
		params := &acm.DescribeCertificateInput{
			CertificateArn: aws.String(certificate_arn),
		}

		resp, err := acmconn.DescribeCertificate(params)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing certificate: %s", err))
		}

		if *resp.Certificate.Status != "ISSUED" {
			return resource.RetryableError(fmt.Errorf("Expected certificate to be issued but was in state %s", *resp.Certificate.Status))
		}

		log.Printf("[INFO] ACM Certificate validation for %s done, certificate was issued", certificate_arn)
		return resource.NonRetryableError(resourceAwsAcmCertificateValidationRead(d, meta))
	})
}

func resourceAwsAcmCertificateValidationRead(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	params := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
	}

	resp, err := acmconn.DescribeCertificate(params)

	if err != nil {
		return fmt.Errorf("Error describing certificate: %s", err)
	}

	if *resp.Certificate.Status != "ISSUED" {
		log.Printf("[INFO] Certificate status not issued, was %s, tainting validation", *resp.Certificate.Status)
		d.SetId("")
	} else {
		d.SetId((*resp.Certificate.IssuedAt).String())
	}
	return nil
}

func resourceAwsAcmCertificateValidationDelete(d *schema.ResourceData, meta interface{}) error {
	// No need to do anything, certificate will be deleted when acm_certificate is deleted
	d.SetId("")
	return nil
}
