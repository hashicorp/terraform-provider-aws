package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsAcmCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmCertificateRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func describeCertificateByArn(arn *string, conn *acm.ACM) (*acm.CertificateDetail, error) {
	input := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(*arn),
	}
	log.Printf("[DEBUG] Describing ACM Certificate: %s", input)
	output, err := conn.DescribeCertificate(input)
	return output.Certificate, err
}

func dataSourceAwsAcmCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn

	params := &acm.ListCertificatesInput{}
	target := d.Get("domain")
	statuses, ok := d.GetOk("statuses")
	if ok {
		statusStrings := statuses.([]interface{})
		params.CertificateStatuses = expandStringList(statusStrings)
	} else {
		params.CertificateStatuses = []*string{aws.String("ISSUED")}
	}

	var arns []*string
	log.Printf("[DEBUG] Reading ACM Certificate: %s", params)
	err := conn.ListCertificatesPages(params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if *cert.DomainName == target {
				arns = append(arns, cert.CertificateArn)
			}
		}

		return true
	})
	if err != nil {
		return fmt.Errorf("Error listing certificates: %q", err)
	}

	if len(arns) == 0 {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	filterMostRecent := d.Get("most_recent").(bool)
	filterTypes, filterTypesOk := d.GetOk("types")

	var matchedCertificate *acm.CertificateDetail

	if !filterMostRecent && !filterTypesOk {
		if len(arns) > 1 {
			// Multiple certificates have been found and no flags set. Error
			return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
		}
		// Only 1 certificate has been found and no flags set. Get details and return it.
		matchedCertificate, err = describeCertificateByArn(arns[0], conn)
		if err != nil {
			return fmt.Errorf("Error describing ACM certificate: %q", err)
		}
	} else {
		typesStrings := expandStringList(filterTypes.([]interface{}))

		for _, arn := range arns {
			certificate, err := describeCertificateByArn(arn, conn)
			if err != nil {
				return fmt.Errorf("Error describing ACM certificate: %q", err)
			}
			if filterTypesOk {
				for _, certType := range typesStrings {
					if *certificate.Type == *certType {
						// We do not have a candidate certificate
						if matchedCertificate == nil {
							matchedCertificate = certificate
							continue
						}
						// At this point, we already have a candidate certificate
						// Check if we are filtering by most recent and update if necessary
						if filterMostRecent && (*certificate.NotBefore).After(*matchedCertificate.NotBefore) {
							matchedCertificate = certificate
							break
						}
						// Now we have multiple candidate certificates and we only allow one certificate
						return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
					}
				}
				continue
			}
			// At this point, we already have a candidate certificate
			// Check if we are filtering by most recent and update if necessary
			if filterMostRecent {
				// We do not have a candidate certificate
				if matchedCertificate == nil {
					matchedCertificate = certificate
					continue
				}
				if (*certificate.NotBefore).After(*matchedCertificate.NotBefore) {
					matchedCertificate = certificate
				}
				continue
			}
			// Now we have multiple candidate certificates and we only allow one certificate
			return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
		}
	}

	if matchedCertificate == nil {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("arn", matchedCertificate.CertificateArn)

	return nil
}
