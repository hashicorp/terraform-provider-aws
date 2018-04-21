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
			},
			"domain_validation_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"validation_domain": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"resource_record_name": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "Use `certificate_details[0].resource_record_name` instead",
						},
						"resource_record_type": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "Use `certificate_details[0].resource_record_type` instead",
						},
						"resource_record_value": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "Use `certificate_details[0].resource_record_value` instead",
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_details": {
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
						"validation_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"validation_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"validation_emails": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"validation_emails": {
				Type:       schema.TypeList,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Deprecated: "Use `certificate_details[0].validation_emails` instead",
			},
		},
	}
}

func resourceAwsAcmCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	domainName := d.Get("domain_name").(string)
	params := &acm.RequestCertificateInput{
		DomainName:       aws.String(domainName),
		ValidationMethod: aws.String(d.Get("validation_method").(string)),
	}

	sans, ok := d.GetOk("subject_alternative_names")
	if ok {
		sanStrings := sans.([]interface{})
		params.SubjectAlternativeNames = expandStringList(sanStrings)
	}

	value, ok := d.GetOk("domain_validation_options")

	if ok {
		var domainValidationOptions []*acm.DomainValidationOption
		for _, o := range value.([]interface{}) {
			x := o.(map[string]interface{})
			dn := x["domain_name"].(string)
			vd := x["validation_domain"].(string)
			domainValidationOption := &acm.DomainValidationOption{
				DomainName:       &dn,
				ValidationDomain: &vd,
			}
			domainValidationOptions = append(domainValidationOptions, domainValidationOption)
		}
		params.SetDomainValidationOptions(domainValidationOptions)
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

		if err != nil {
			if isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error describing certificate: %s", err))
		}

		if err := d.Set("domain_name", resp.Certificate.DomainName); err != nil {
			return resource.NonRetryableError(err)
		}

		if err := d.Set("arn", resp.Certificate.CertificateArn); err != nil {
			return resource.NonRetryableError(err)
		}

		if err := d.Set("subject_alternative_names", cleanUpSubjectAlternativeNames(resp.Certificate)); err != nil {
			return resource.NonRetryableError(err)
		}

		certificateDetails, err := convertCertificateDetails(resp.Certificate)

		if err != nil {
			return resource.RetryableError(err)
		}

		if len(certificateDetails) < 1 {
			return resource.NonRetryableError(fmt.Errorf("Error getting certificate details"))
		}

		if err := d.Set("certificate_details", certificateDetails); err != nil {
			return resource.NonRetryableError(err)
		}

		d.Set("validation_method", certificateDetails[0]["validation_method"])

		params := &acm.ListTagsForCertificateInput{
			CertificateArn: aws.String(d.Id()),
		}

		tagResp, err := acmconn.ListTagsForCertificate(params)
		if err := d.Set("tags", tagsToMapACM(tagResp.Tags)); err != nil {
			return resource.NonRetryableError(err)
		}

		//support for deprecated attributes
		d.Set("validation_emails", certificateDetails[0]["validation_emails"])

		value, ok := d.GetOk("domain_validation_options")
		var domainValidationOptionsInput []interface{}
		if ok {
			domainValidationOptionsInput = value.([]interface{})
			//pad until it is the same size as certificate details
			for len(domainValidationOptionsInput) < len(certificateDetails) {
				domainValidationOptionsInput = append(domainValidationOptionsInput, make(map[string]interface{}, 1))
			}
		} else {
			for i := 0; i < len(certificateDetails); i++ {
				domainValidationOptionsInput = append(domainValidationOptionsInput, make(map[string]interface{}, 1))
			}
		}
		for i, v := range domainValidationOptionsInput {
			validationOption := v.(map[string]interface{})
			validationOption["domain_name"] = certificateDetails[i]["domain_name"]
			validationOption["resource_record_name"] = certificateDetails[i]["resource_record_name"]
			validationOption["resource_record_type"] = certificateDetails[i]["resource_record_type"]
			validationOption["resource_record_value"] = certificateDetails[i]["resource_record_value"]
		}

		if err := d.Set("domain_validation_options", domainValidationOptionsInput); err != nil {
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

func convertCertificateDetails(certificate *acm.CertificateDetail) ([]map[string]interface{}, error) {
	var certificateDetails []map[string]interface{}

	if *certificate.Type == acm.CertificateTypeAmazonIssued {
		for _, o := range certificate.DomainValidationOptions {
			var validationOption map[string]interface{}
			if o.ResourceRecord != nil {
				validationOption = map[string]interface{}{
					"domain_name":           *o.DomainName,
					"resource_record_name":  *o.ResourceRecord.Name,
					"resource_record_type":  *o.ResourceRecord.Type,
					"resource_record_value": *o.ResourceRecord.Value,
					"validation_method":     *o.ValidationMethod,
				}

			} else if o.ValidationEmails != nil && len(o.ValidationEmails) > 0 {
				var validationEmails []string
				for _, email := range o.ValidationEmails {
					validationEmails = append(validationEmails, *email)
				}
				validationOption = map[string]interface{}{
					"domain_name":       *o.DomainName,
					"validation_emails": validationEmails,
					"validation_method": "EMAIL",
				}
			} else {
				return nil, fmt.Errorf("Validation options not yet updated. Need to retry: %#v", o)
			}

			if o.ValidationDomain != nil {
				validationOption["validation_domain"] = *o.ValidationDomain
			}
			certificateDetails = append(certificateDetails, validationOption)
		}
	}
	return certificateDetails, nil
}

func resourceAwsAcmCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	log.Printf("[INFO] Deleting ACM Certificate: %s", d.Id())

	params := &acm.DeleteCertificateInput{
		CertificateArn: aws.String(d.Id()),
	}

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := acmconn.DeleteCertificate(params)
		if err != nil {
			if isAWSErr(err, acm.ErrCodeResourceInUseException, "") {
				log.Printf("[WARN] Conflict deleting certificate in use: %s, retrying", err.Error())
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil && !isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
		return fmt.Errorf("Error deleting certificate: %s", err)
	}

	d.SetId("")
	return nil
}
