package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderInfrastructureConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"datecreated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dateupdated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_profile_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"s3_key_prefix": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sns_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsImageBuilderInfrastructureConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	componentArn := d.Get("arn").(string)

	params := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(componentArn),
	}

	resp, err := conn.GetInfrastructureConfiguration(params)

	if err != nil {
		return fmt.Errorf("Error retrieving Component: %s", err)
	}

	d.SetId(*resp.InfrastructureConfiguration.Arn)
	d.Set("datecreated", resp.InfrastructureConfiguration.DateCreated)
	d.Set("dateupdated", resp.InfrastructureConfiguration.DateUpdated)
	d.Set("date_created", resp.InfrastructureConfiguration.DateCreated)
	d.Set("description", resp.InfrastructureConfiguration.Description)
	d.Set("instance_profile_name", resp.InfrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", resp.InfrastructureConfiguration.InstanceTypes)
	d.Set("key_pair", resp.InfrastructureConfiguration.KeyPair)
	d.Set("logging", resp.InfrastructureConfiguration.Logging)
	d.Set("name", resp.InfrastructureConfiguration.Name)
	d.Set("security_group_ids", resp.InfrastructureConfiguration.SecurityGroupIds)
	d.Set("sns_topic_arn", resp.InfrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", resp.InfrastructureConfiguration.SubnetId)
	d.Set("terminate_instance_on_failure", resp.InfrastructureConfiguration.TerminateInstanceOnFailure)
	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(resp.InfrastructureConfiguration.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	return nil
}
