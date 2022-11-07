package apigateway

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// cloudFrontRoute53ZoneID defines the route 53 zone ID for CloudFront. This
// is used to set the zone_id attribute.
const cloudFrontRoute53ZoneID = "Z2FDTNDATAQYW2"

func DataSourceDomainName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDomainNameRead,
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
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceDomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &apigateway.GetDomainNameInput{}

	if v, ok := d.GetOk("domain_name"); ok {
		input.DomainName = aws.String(v.(string))
	}

	domainName, err := conn.GetDomainName(input)

	if err != nil {
		return fmt.Errorf("error getting API Gateway Domain Name: %w", err)
	}

	d.SetId(aws.StringValue(domainName.DomainName))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("certificate_arn", domainName.CertificateArn)
	d.Set("certificate_name", domainName.CertificateName)

	if domainName.CertificateUploadDate != nil {
		d.Set("certificate_upload_date", domainName.CertificateUploadDate.Format(time.RFC3339))
	}

	d.Set("cloudfront_domain_name", domainName.DistributionDomainName)
	d.Set("cloudfront_zone_id", cloudFrontRoute53ZoneID)
	d.Set("domain_name", domainName.DomainName)

	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(domainName.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %w", err)
	}

	d.Set("regional_certificate_arn", domainName.RegionalCertificateArn)
	d.Set("regional_certificate_name", domainName.RegionalCertificateName)
	d.Set("regional_domain_name", domainName.RegionalDomainName)
	d.Set("regional_zone_id", domainName.RegionalHostedZoneId)
	d.Set("security_policy", domainName.SecurityPolicy)

	if err := d.Set("tags", KeyValueTags(domainName.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
