package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAcmReqCertRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAcmReqCertCreate,
		Read:   resourceAcmReqCertRead,
		Delete: resourceAcmReqCertDelete,
		Update: resourceAcmReqCertUpdate,

		Schema: map[string]*schema.Schema{
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"validation_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "DNS",
			},
			"idempotency_token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"subject_alternatives_names": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"validation_options": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"validation_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAcmReqCertCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn

	input := &acm.RequestCertificateInput{
		DomainName:       aws.String(d.Get("domain_name").(string)),
		ValidationMethod: aws.String(acm.ValidationMethodDns),
	}

	if value, ok := d.GetOk("idempotency_token"); ok {
		input.SetIdempotencyToken(value.(string))
	}

	if value, ok := d.GetOk("subject_alternatives_names"); ok {
		input.SetSubjectAlternativeNames(expandStringSet(value.(*schema.Set)))
	}

	if d.Get("validation_method").(string) == "EMAIL" {
		input.SetValidationMethod(acm.ValidationMethodEmail)
	}

	if value, ok := d.GetOk("validation_options"); ok {
		options := value.(*schema.Set).List()

		for _, value := range options {
			validation_options := &acm.DomainValidationOption{
				DomainName:       aws.String(value.(*schema.ResourceData).Get("domain_name").(string)),
				ValidationDomain: aws.String(value.(*schema.ResourceData).Get("validation_domain").(string)),
			}
			input.DomainValidationOptions = append(input.DomainValidationOptions, validation_options)
		}
	}

	log.Printf("[DEBUG] RequestCertificate input data: %v", input)

	reqOutput, err := conn.RequestCertificate(input)

	if err != nil {
		return err
	}

	arn := *reqOutput.CertificateArn
	d.Set("arn", arn)
	d.Set("idempotency_token", fmt.Sprintf("%d", time.Now().Unix()+3600))
	d.SetId(resourceAcmGetIdFromArn(arn))

	return nil
}

func resourceAcmReqCertRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn

	newarn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "acm",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("certificate/%s", d.Id()),
	}

	input := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(newarn.String()),
	}

	description, err := conn.DescribeCertificate(input)

	if err != nil {
		if acmerr, ok := err.(awserr.Error); ok {
			if acmerr.Code() == "ResourceNotFoundException" {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	certificate := description.Certificate
	d.SetId(resourceAcmGetIdFromArn(*certificate.CertificateArn))
	d.Set("arn", *certificate.CertificateArn)
	d.Set("domain_name", *certificate.DomainName)
	d.Set("subject_alternatives_names", certificate.SubjectAlternativeNames)
	d.Set("validation_options", certificate.DomainValidationOptions)
	return nil
}

func resourceAcmReqCertDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn

	input := &acm.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get("arn").(string)),
	}

	_, err := conn.DeleteCertificate(input)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceAcmReqCertUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn

	time_token, err := strconv.ParseInt(d.Get("idempotency_token").(string), 10, 64)

	if err != nil {
		log.Printf("[ERROR] Can't convert idempotency_token string to int64. Error: %s", err)
		return err
	}

	if d.HasChange("domain_name") || time_token > time.Now().Unix() {
		if err := resourceAcmReqCertDelete(d, meta); err != nil {
			return err
		}
		return resourceAcmReqCertCreate(d, meta)
	}

	input := &acm.RequestCertificateInput{
		DomainName:       aws.String(d.Get("domain_name").(string)),
		ValidationMethod: aws.String(acm.ValidationMethodDns),
	}

	if d.HasChange("subject_alternatives_names") {
		input.SetSubjectAlternativeNames(expandStringSet(d.Get("subject_alternatives_names").(*schema.Set)))
	}

	reqOutput, err := conn.RequestCertificate(input)

	if err != nil {
		return err
	}

	d.Set("arn", reqOutput.CertificateArn)

	return nil
}

func resourceAcmGetIdFromArn(inarn string) string {
	splited := strings.Split(inarn, "/")

	if len(splited) >= 1 {
		return splited[1]
	}

	return inarn
}
