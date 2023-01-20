package acm

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,

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
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(acm.KeyAlgorithm_Values(), false),
				},
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"statuses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ACMConn()
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
	err := conn.ListCertificatesPagesWithContext(ctx, params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if aws.StringValue(cert.DomainName) == target {
				arns = append(arns, cert.CertificateArn)
			}
		}

		return true
	})

	if err != nil {
		return diag.Errorf("listing ACM Certificates: %s", err)
	}

	if len(arns) == 0 {
		return diag.Errorf("no certificate for domain %q found in this Region", target)
	}

	filterMostRecent := d.Get("most_recent").(bool)
	filterTypes, filterTypesOk := d.GetOk("types")

	var matchedCertificate *acm.CertificateDetail

	if !filterMostRecent && !filterTypesOk && len(arns) > 1 {
		// Multiple certificates have been found and no additional filtering set
		return diag.Errorf("multiple certificates for domain %q found in this Region", target)
	}

	typesStrings := flex.ExpandStringList(filterTypes.([]interface{}))

	for _, arn := range arns {
		arn := aws.StringValue(arn)
		output, err := conn.DescribeCertificateWithContext(ctx, &acm.DescribeCertificateInput{
			CertificateArn: aws.String(arn),
		})

		if err != nil {
			return diag.Errorf("reading ACM Certificate (%s): %s", arn, err)
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
							return diag.FromErr(err)
						}
						break
					}
					// Now we have multiple candidate certificates and we only allow one certificate
					return diag.Errorf("multiple certificates for domain %q found in this Region", target)
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
				return diag.FromErr(err)
			}
			continue
		}
		// Now we have multiple candidate certificates and we only allow one certificate
		return diag.Errorf("multiple certificates for domain %q found in this Region", target)
	}

	if matchedCertificate == nil {
		return diag.Errorf("no certificate for domain %q found in this Region", target)
	}

	// Get the certificate data if the status is issued
	var certOutput *acm.GetCertificateOutput
	if aws.StringValue(matchedCertificate.Status) == acm.CertificateStatusIssued {
		arn := aws.StringValue(matchedCertificate.CertificateArn)
		certOutput, err = conn.GetCertificateWithContext(ctx, &acm.GetCertificateInput{
			CertificateArn: aws.String(arn),
		})

		if err != nil {
			return diag.Errorf("reading ACM Certificate (%s): %s", arn, err)
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

	tags, err := ListTags(ctx, conn, aws.StringValue(matchedCertificate.CertificateArn))

	if err != nil {
		return diag.Errorf("listing tags for ACM Certificate (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
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
