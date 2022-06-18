package acm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCertificateRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
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
			"statuses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						acm.KeyAlgorithm_Values(), false,
					),
				},
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &acm.ListCertificatesInput{}

	if v := d.Get("key_types").(*schema.Set); v.Len() > 0 {
		params.Includes = &acm.Filters{
			KeyTypes: flex.ExpandStringSet(v),
		}
	}

	target := d.Get("domain")
	statuses, ok := d.GetOk("statuses")
	if ok {
		statusStrings := statuses.([]interface{})
		params.CertificateStatuses = flex.ExpandStringList(statusStrings)
	} else {
		params.CertificateStatuses = []*string{aws.String(acm.CertificateStatusIssued)}
	}

	var arns []*string
	log.Printf("[DEBUG] Reading ACM Certificate: %s", params)
	err := conn.ListCertificatesPages(params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if aws.StringValue(cert.DomainName) == target {
				arns = append(arns, cert.CertificateArn)
			}
		}

		return true
	})
	if err != nil {
		return fmt.Errorf("Error listing certificates: %w", err)
	}

	if len(arns) == 0 {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	filterMostRecent := d.Get("most_recent").(bool)
	filterTypes, filterTypesOk := d.GetOk("types")

	var matchedCertificate *acm.CertificateDetail

	if !filterMostRecent && !filterTypesOk && len(arns) > 1 {
		// Multiple certificates have been found and no additional filtering set
		return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
	}

	typesStrings := flex.ExpandStringList(filterTypes.([]interface{}))

	for _, arn := range arns {
		var err error

		input := &acm.DescribeCertificateInput{
			CertificateArn: aws.String(*arn),
		}
		log.Printf("[DEBUG] Describing ACM Certificate: %s", input)
		output, err := conn.DescribeCertificate(input)
		if err != nil {
			return fmt.Errorf("Error describing ACM certificate: %w", err)
		}
		certificate := output.Certificate

		if filterTypesOk {
			for _, certType := range typesStrings {
				if aws.StringValue(certificate.Type) == aws.StringValue(certType) {
					// We do not have a candidate certificate
					if matchedCertificate == nil {
						matchedCertificate = certificate
						break
					}
					// At this point, we already have a candidate certificate
					// Check if we are filtering by most recent and update if necessary
					if filterMostRecent {
						matchedCertificate, err = mostRecentCertificate(certificate, matchedCertificate)
						if err != nil {
							return err
						}
						break
					}
					// Now we have multiple candidate certificates and we only allow one certificate
					return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
				}
			}
			continue
		}
		// We do not have a candidate certificate
		if matchedCertificate == nil {
			matchedCertificate = certificate
			continue
		}
		// At this point, we already have a candidate certificate
		// Check if we are filtering by most recent and update if necessary
		if filterMostRecent {
			matchedCertificate, err = mostRecentCertificate(certificate, matchedCertificate)
			if err != nil {
				return err
			}
			continue
		}
		// Now we have multiple candidate certificates and we only allow one certificate
		return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
	}

	if matchedCertificate == nil {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	// Get the certificate data if the status is issued
	var certOutput *acm.GetCertificateOutput
	if aws.StringValue(matchedCertificate.Status) == acm.CertificateStatusIssued {
		getCertInput := acm.GetCertificateInput{
			CertificateArn: matchedCertificate.CertificateArn,
		}
		certOutput, err = conn.GetCertificate(&getCertInput)
		if err != nil {
			return fmt.Errorf("error getting ACM certificate (%s): %w", aws.StringValue(matchedCertificate.CertificateArn), err)
		}
	}
	if certOutput != nil {
		d.Set("certificate", certOutput.Certificate)
		d.Set("certificate_chain", certOutput.CertificateChain)
	} else {
		d.Set("certificate", nil)
		d.Set("certificate_chain", nil)
	}

	d.SetId(aws.StringValue(matchedCertificate.CertificateArn))
	d.Set("arn", matchedCertificate.CertificateArn)
	d.Set("status", matchedCertificate.Status)
	tags, err := ListTags(conn, aws.StringValue(matchedCertificate.CertificateArn))
	if err != nil {
		return fmt.Errorf("error listing tags for ACM Certificate (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func mostRecentCertificate(i, j *acm.CertificateDetail) (*acm.CertificateDetail, error) {
	if aws.StringValue(i.Status) != aws.StringValue(j.Status) {
		return nil, fmt.Errorf("most_recent filtering on different ACM certificate statues is not supported")
	}
	// Cover IMPORTED and ISSUED AMAZON_ISSUED certificates
	if aws.StringValue(i.Status) == acm.CertificateStatusIssued {
		if aws.TimeValue(i.NotBefore).After(aws.TimeValue(j.NotBefore)) {
			return i, nil
		}
		return j, nil
	}
	// Cover non-ISSUED AMAZON_ISSUED certificates
	if aws.TimeValue(i.CreatedAt).After(aws.TimeValue(j.CreatedAt)) {
		return i, nil
	}
	return j, nil
}
