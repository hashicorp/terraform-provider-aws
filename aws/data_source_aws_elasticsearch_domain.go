package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/structure"
)

func dataSourceAwsElasticSearchDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticSearchDomainRead,

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"advanced_options": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
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
			"kibana_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
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
					},
				},
			},
			"encryption_at_rest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"node_to_node_encryption": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dedicated_master_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"dedicated_master_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_awareness_enabled": {
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
			"vpc_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							//Set:      schema.HashString,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloudwatch_log_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_options": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"created": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"processing": {
				Type:     schema.TypeString,
				Computed: true,
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
		return fmt.Errorf("your query returned no results")
	}

	ds := resp.DomainStatus

	d.SetId(*ds.ARN)

	if ds.AccessPolicies != nil && *ds.AccessPolicies != "" {
		policies, err := structure.NormalizeJsonString(*ds.AccessPolicies)
		if err != nil {
			return errwrap.Wrapf("access policies contain an invalid JSON: {{err}}", err)
		}
		d.Set("access_policies", policies)
	}

	err = d.Set("advanced_options", pointersMapToStringList(ds.AdvancedOptions))
	if err != nil {
		return err
	}

	d.Set("arn", *ds.ARN)
	d.Set("domain_id", ds.DomainId)
	d.Set("endpoint", ds.Endpoint)
	d.Set("kibana_endpoint", getKibanaEndpoint(d))

	err = d.Set("ebs_options", flattenESEBSOptions(ds.EBSOptions))
	if err != nil {
		return err
	}

	err = d.Set("encryption_at_rest", flattenESEncryptAtRestOptions(ds.EncryptionAtRestOptions))
	if err != nil {
		return err
	}

	err = d.Set("node_to_node_encryption", flattenESNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions))
	if err != nil {
		return err
	}

	err = d.Set("cluster_config", flattenESClusterConfig(ds.ElasticsearchClusterConfig))
	if err != nil {
		return err
	}

	if ds.SnapshotOptions != nil {
		m := map[string]interface{}{}

		m["automated_snapshot_start_hour"] = *ds.SnapshotOptions.AutomatedSnapshotStartHour

		d.Set("snapshot_options", []map[string]interface{}{m})
	}

	if ds.VPCOptions != nil {
		err = d.Set("vpc_options", flattenESVPCDerivedInfo(ds.VPCOptions))
		if err != nil {
			return err
		}

		endpoints := pointersMapToStringList(ds.Endpoints)
		err = d.Set("endpoint", endpoints["vpc"])
		if err != nil {
			return err
		}
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return fmt.Errorf("%q: Elasticsearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set("endpoint", aws.StringValue(ds.Endpoint))
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return fmt.Errorf("%q: Elasticsearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if ds.LogPublishingOptions != nil {
		m := make([]map[string]interface{}, 0)
		for k, val := range ds.LogPublishingOptions {
			mm := map[string]interface{}{}
			mm["log_type"] = k
			if val.CloudWatchLogsLogGroupArn != nil {
				mm["cloudwatch_log_group_arn"] = aws.StringValue(val.CloudWatchLogsLogGroupArn)
			}
			mm["enabled"] = aws.BoolValue(val.Enabled)
			m = append(m, mm)
		}
		d.Set("log_publishing_options", m)
	}

	d.Set("elasticsearch_version", ds.ElasticsearchVersion)

	err = d.Set("cognito_options", flattenESCognitoOptions(ds.CognitoOptions))
	if err != nil {
		return err
	}

	d.Set("created", ds.Created)
	d.Set("deleted", ds.Deleted)

	d.Set("processing", ds.Processing)

	tagResp, err := esconn.ListTags(&elasticsearchservice.ListTagsInput{
		ARN: ds.ARN,
	})

	if err != nil {
		log.Printf("[DEBUG] Error retrieving tags for ARN: %s", *ds.ARN)
	}

	d.Set("tags", tagsToMapElasticsearchService(tagResp.TagList))

	return nil
}
