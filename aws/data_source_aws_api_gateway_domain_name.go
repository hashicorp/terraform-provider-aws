package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsApiGatewayDomainName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsApiGatewayDomainNameRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"regional_certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_certificate_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsApiGatewayDomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	targetDomainName := d.Get("domain_name").(string)
	log.Printf("[DEBUG] Reading API Gateway Domain Name %s", targetDomainName)
	domainName, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
		DomainName: aws.String(targetDomainName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == apigateway.ErrCodeNotFoundException {
			return fmt.Errorf("API Gateway Domain Name (%s) not found", targetDomainName)
		}

		return err
	}

	d.SetId(*domainName.DomainName)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("certificate_arn", domainName.CertificateArn)
	d.Set("certificate_name", domainName.CertificateName)
	if err := d.Set("certificate_upload_date", domainName.CertificateUploadDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting certificate_upload_date: %s", err)
	}
	d.Set("cloudfront_domain_name", domainName.DistributionDomainName)
	d.Set("cloudfront_zone_id", cloudFrontRoute53ZoneID)
	d.Set("domain_name", domainName.DomainName)
	d.Set("security_policy", domainName.SecurityPolicy)

	if err := d.Set("endpoint_configuration", flattenApiGatewayEndpointConfiguration(domainName.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %s", err)
	}

	d.Set("regional_certificate_arn", domainName.RegionalCertificateArn)
	d.Set("regional_certificate_name", domainName.RegionalCertificateName)
	d.Set("regional_domain_name", domainName.RegionalDomainName)
	d.Set("regional_zone_id", domainName.RegionalHostedZoneId)

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(domainName.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
