package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsElasticSearchDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticSearchDomainRead,

		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_policies": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"processing": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"zone_awareness_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"dedicated_master_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dedicated_master_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"snapshot_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automated_snapshot_start_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"advanced_options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsElasticSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	esconn := meta.(*AWSClient).esconn

	req := &elasticsearchservice.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	resp, err := esconn.DescribeElasticsearchDomain(req)
	if err != nil {
		return err
	}

	if resp.DomainStatus == nil {
		return fmt.Errorf("Your query returned no results.  Please change your search criteria and try again.")
	}

	ds := resp.DomainStatus

	d.SetId(*ds.ARN)

	d.Set("domain_id", ds.DomainId)
	d.Set("endpoint", ds.Endpoint)
	d.Set("created", ds.Created)
	d.Set("deleted", ds.Deleted)

	if ds.AccessPolicies != nil && *ds.AccessPolicies != "" {
		policies, err := normalizeJsonString(*ds.AccessPolicies)
		if err != nil {
			return errwrap.Wrapf("access policies contain an invalid JSON: {{err}}", err)
		}
		d.Set("access_policies", policies)
	}

	d.Set("processing", ds.Processing)
	d.Set("elasticsearch_version", ds.ElasticsearchVersion)
	arn := *ds.ARN
	d.Set("arn", arn)

	err = d.Set("advanced_options", pointersMapToStringList(ds.AdvancedOptions))
	if err != nil {
		return err
	}

	err = d.Set("cluster_config", flattenESClusterConfig(ds.ElasticsearchClusterConfig))
	if err != nil {
		return err
	}

	err = d.Set("ebs_options", flattenESEBSOptions(ds.EBSOptions))
	if err != nil {
		return err
	}

	if ds.SnapshotOptions != nil {
		m := map[string]interface{}{}

		m["automated_snapshot_start_hour"] = *ds.SnapshotOptions.AutomatedSnapshotStartHour

		d.Set("snapshot_options", []map[string]interface{}{m})
	}

	tagResp, err := esconn.ListTags(&elasticsearchservice.ListTagsInput{
		ARN: &arn,
	})

	if err != nil {
		log.Printf("[DEBUG] Error retrieving tags for ARN: %s", arn)
	}

	d.Set("tags", tagsToMapElasticsearchService(tagResp.TagList))

	return nil
}
