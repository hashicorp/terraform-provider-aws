package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/errwrap"
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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

type arnData struct {
	arn       string
	notBefore *time.Time
}

func describeCertificate(arn *arnData, conn *acm.ACM) (*acm.DescribeCertificateOutput, error) {
	params := &acm.DescribeCertificateInput{}
	params.CertificateArn = &arn.arn

	description, err := conn.DescribeCertificate(params)
	if err != nil {
		return nil, err
	}

	return description, nil
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

	var arns []*arnData
	log.Printf("[DEBUG] Reading ACM Certificate: %s", params)
	err := conn.ListCertificatesPages(params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if *cert.DomainName == target {
				arns = append(arns, &arnData{*cert.CertificateArn, nil})
			}
		}

		return true
	})
	if err != nil {
		return errwrap.Wrapf("Error listing certificates: {{err}}", err)
	}

	// filter based on certificate type (imported or aws-issued)
	types, ok := d.GetOk("types")
	if ok {
		typesStrings := expandStringList(types.([]interface{}))
		var matchedArns []*arnData
		for _, arn := range arns {
			description, err := describeCertificate(arn, conn)
			if err != nil {
				return errwrap.Wrapf("Error describing certificates: {{err}}", err)
			}

			for _, certType := range typesStrings {
				if *description.Certificate.Type == *certType {
					matchedArns = append(
						matchedArns,
						&arnData{arn.arn, description.Certificate.NotBefore},
					)
					break
				}
			}
		}

		arns = matchedArns
	}

	if len(arns) == 0 {
		return fmt.Errorf("No certificate for domain %q found in this region.", target)
	}

	if len(arns) > 1 {
		// Get most recent sorting by notBefore date. Notice that createdAt field is only valid
		// for ACM issued certificates but not for imported ones so in a mixed scenario only
		// fields extracted from the certificate are valid.
		_, ok = d.GetOk("most_recent")
		if ok {
			mr := arns[0]
			if mr.notBefore == nil {
				description, err := describeCertificate(mr, conn)
				if err != nil {
					return errwrap.Wrapf("Error describing certificates: {{err}}", err)
				}

				mr.notBefore = description.Certificate.NotBefore
			}
			for _, arn := range arns[1:] {
				if arn.notBefore == nil {
					description, err := describeCertificate(arn, conn)
					if err != nil {
						return errwrap.Wrapf("Error describing certificates: {{err}}", err)
					}

					arn.notBefore = description.Certificate.NotBefore
				}

				if arn.notBefore.After(*mr.notBefore) {
					mr = arn
				}
			}

			arns = []*arnData{mr}
		} else {
			return fmt.Errorf("Multiple certificates for domain %q found in this region.", target)
		}
	}

	d.SetId(time.Now().UTC().String())
	d.Set("arn", arns[0].arn)

	return nil
}
