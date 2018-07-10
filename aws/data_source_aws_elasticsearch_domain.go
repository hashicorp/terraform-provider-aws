package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

func dataSourceAwsElasticsearchDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticSearchDomainRead,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsElasticSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	domainName := aws.String(d.Get("domain_name").(string))
	req := &elasticsearch.DescribeElasticsearchDomainsInput{
		DomainNames: []*string{domainName},
	}

	log.Printf("[DEBUG] Reading ElasticSearch Domains.")
	d.SetId(time.Now().UTC().String())

	resp, err := conn.DescribeElasticsearchDomains(req)
	if err != nil {
		return err
	}

	if len(resp.DomainStatusList) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}
	if len(resp.DomainStatusList) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	domain := resp.DomainStatusList[0]

	d.SetId(*domain.DomainId)
	d.Set("domain_id", domain.DomainId)
	d.Set("endpoint", domain.Endpoint)

	arn, err := buildElasticserchARN(*domainName, meta.(*AWSClient).accountid, meta.(*AWSClient).region)
	if err != nil {
		log.Printf("[DEBUG] Error building ARN for ElastiCache Cluster %s", *domain.DomainId)
	}
	d.Set("arn", arn)

	tagResp, err := conn.ListTags(&elasticsearch.ListTagsInput{
		ARN: aws.String(arn),
	})

	if err != nil {
		log.Printf("[DEBUG] Error retrieving tags for ARN: %s", arn)
	}

	d.Set("tags", tagsToMapElasticsearchService(tagResp.TagList))

	return nil
}

func buildElasticserchARN(domain, accountid, region string) (string, error) {
	if domain == "" {
		return "", fmt.Errorf("Unable to construct Elasticsearch ARN because of missing domain")
	}
	if accountid == "" {
		return "", fmt.Errorf("Unable to construct Elasticsearch ARN because of missing accountid")
	}
	arn := fmt.Sprintf("arn:aws:es:%s:%s:domain/%s", region, accountid, domain)
	return arn, nil
}
