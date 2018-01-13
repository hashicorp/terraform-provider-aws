package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsAcmCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmCertificateRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"statuses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_until_present": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wait_until_present_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "45m",
			},
		},
	}
}

func dataSourceAwsAcmCertificateRead(d *schema.ResourceData, meta interface{}) error {
	if d.Get("wait_until_present").(bool) {

		timeout, err := time.ParseDuration(d.Get("wait_until_present_timeout").(string))
		if err != nil {
			return err
		}
		return resource.Retry(timeout, func() *resource.RetryError {
			return dataSourceAwsAcmGetCertificate(d, meta)
		})
	} else {
		return dataSourceAwsAcmGetCertificate(d, meta).Err
	}
}

func dataSourceAwsAcmGetCertificate(d *schema.ResourceData, meta interface{}) *resource.RetryError {
	conn := meta.(*AWSClient).acmconn

	params := &acm.ListCertificatesInput{}

	targetDomain, _ := d.GetOk("domain")
	targetArn, _ := d.GetOk("arn")

	if targetArn == nil && targetDomain == nil {
		return resource.NonRetryableError(fmt.Errorf("Need to specify either domain or arn"))
	}

	statuses, ok := d.GetOk("statuses")
	if ok {
		statusStrings := statuses.([]interface{})
		params.CertificateStatuses = expandStringList(statusStrings)
	} else {
		params.CertificateStatuses = []*string{aws.String("ISSUED")}
	}

	var arns []string
	var domains []string

	log.Printf("[DEBUG] Reading ACM Certificate: %s", params)
	err := conn.ListCertificatesPages(params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if targetDomain != nil && *cert.DomainName == targetDomain {
				arns = append(arns, *cert.CertificateArn)
				domains = append(domains, *cert.DomainName)
			}
			if targetArn != nil && *cert.CertificateArn == targetArn {
				arns = append(arns, *cert.CertificateArn)
				domains = append(domains, *cert.DomainName)
			}
		}

		return true
	})
	if err != nil {
		return resource.NonRetryableError(errwrap.Wrapf("Error describing certificates: {{err}}", err))
	}

	// filter based on certificate type (imported or aws-issued)
	types, ok := d.GetOk("types")
	if ok {
		typesStrings := expandStringList(types.([]interface{}))
		var matchedArns []string
		var matchedDomains []string
		for _, arn := range arns {
			params := &acm.DescribeCertificateInput{}
			params.CertificateArn = &arn

			description, err := conn.DescribeCertificate(params)
			if err != nil {
				return resource.NonRetryableError(errwrap.Wrapf("Error describing certificates: {{err}}", err))
			}

			for _, certType := range typesStrings {
				if *description.Certificate.Type == *certType {
					matchedArns = append(matchedArns, arn)
					matchedDomains = append(domains, *description.Certificate.DomainName)
					break
				}
			}
		}

		arns = matchedArns
		domains = matchedDomains
	}

	targetValue := targetArn
	if targetValue == nil {
		targetValue = targetDomain
	}

	if len(arns) == 0 {
		return resource.RetryableError(fmt.Errorf("No certificate for domain %q found in this region.", targetValue))
	}
	if len(arns) > 1 {
		return resource.NonRetryableError(fmt.Errorf("Multiple certificates for domain %q found in this region.", targetValue))
	}

	d.SetId(time.Now().UTC().String())
	d.Set("arn", arns[0])
	d.Set("domain", domains[0])

	return nil
}
