package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAcmCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmCertificateCreate,
		Read:   resourceAwsAcmCertificateRead,
		Update: resourceAwsAcmCertificateUpdate,
		Delete: resourceAwsAcmCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subject_alternative_names": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"validation_method": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					if v.(string) != acm.ValidationMethodDns {
						errors = append(errors, fmt.Errorf("only validation_method DNS is supported at the moment"))
					}
					return
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_validation_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAcmCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	params := &acm.RequestCertificateInput{
		DomainName:       aws.String(d.Get("domain_name").(string)),
		ValidationMethod: aws.String(acm.ValidationMethodDns),
	}

	sans, ok := d.GetOk("subject_alternative_names")
	if ok {
		sanStrings := sans.([]interface{})
		params.SubjectAlternativeNames = expandStringList(sanStrings)
	}

	log.Printf("[DEBUG] ACM Certificate Request: %#v", params)
	resp, err := acmconn.RequestCertificate(params)

	if err != nil {
		return fmt.Errorf("Error requesting certificate: %s", err)
	}

	d.SetId(*resp.CertificateArn)
	if v, ok := d.GetOk("tags"); ok {
		params := &acm.AddTagsToCertificateInput{
			CertificateArn: resp.CertificateArn,
			Tags:           tagsFromMapACM(v.(map[string]interface{})),
		}
		_, err := acmconn.AddTagsToCertificate(params)

		if err != nil {
			return fmt.Errorf("Error requesting certificate: %s", err)
		}
	}

	return resourceAwsAcmCertificateRead(d, meta)
}

func resourceAwsAcmCertificateRead(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	params := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(d.Id()),
	}

	return resource.Retry(time.Duration(1)*time.Minute, func() *resource.RetryError {
		resp, err := acmconn.DescribeCertificate(params)

		if err != nil && isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
			d.SetId("")
			return nil
		} else if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing certificate: %s", err))
		}

		if *resp.Certificate.Type != acm.CertificateTypeAmazonIssued {
			return resource.NonRetryableError(fmt.Errorf("Certificate has type %s, only AMAZON_ISSUED is supported at the moment", *resp.Certificate.Type))
		}

		d.Set("domain_name", resp.Certificate.DomainName)
		d.Set("arn", resp.Certificate.CertificateArn)
		d.Set("validation_method", resp.Certificate.DomainValidationOptions[0].ValidationMethod)

		if err := d.Set("subject_alternative_names", cleanUpSubjectAlternativeNames(resp.Certificate)); err != nil {
			return resource.NonRetryableError(err)
		}

		domainValidationOptions, err := convertDomainValidationOptions(resp.Certificate.DomainValidationOptions)

		if err != nil {
			return resource.RetryableError(err)
		}

		if err := d.Set("domain_validation_options", domainValidationOptions); err != nil {
			return resource.NonRetryableError(err)
		}

		params := &acm.ListTagsForCertificateInput{
			CertificateArn: aws.String(d.Id()),
		}

		tagResp, err := acmconn.ListTagsForCertificate(params)
		if err := d.Set("tags", tagsToMapACM(tagResp.Tags)); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
}

func resourceAwsAcmCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("tags") {
		acmconn := meta.(*AWSClient).acmconn
		err := setTagsACM(acmconn, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanUpSubjectAlternativeNames(cert *acm.CertificateDetail) []string {
	sans := cert.SubjectAlternativeNames
	vs := make([]string, 0, len(sans)-1)
	for _, v := range sans {
		if *v != *cert.DomainName {
			vs = append(vs, *v)
		}
	}
	return vs

}

func convertDomainValidationOptions(validations []*acm.DomainValidation) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, o := range validations {
		if o.ResourceRecord != nil {
			validationOption := map[string]interface{}{
				"domain_name":           *o.DomainName,
				"resource_record_name":  *o.ResourceRecord.Name,
				"resource_record_type":  *o.ResourceRecord.Type,
				"resource_record_value": *o.ResourceRecord.Value,
			}
			result = append(result, validationOption)
		} else {
			log.Printf("[DEBUG] No resource record found in validation options, need to retry: %#v", o)
			return nil, fmt.Errorf("No resource record found in DNS DomainValidationOptions: %v", o)
		}
	}

	return result, nil
}

func resourceAwsAcmCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	params := &acm.DeleteCertificateInput{
		CertificateArn: aws.String(d.Id()),
	}

	_, err := acmconn.DeleteCertificate(params)

	if err != nil && !isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
		return fmt.Errorf("Error deleting certificate: %s", err)
	}

	d.SetId("")
	return nil
}
