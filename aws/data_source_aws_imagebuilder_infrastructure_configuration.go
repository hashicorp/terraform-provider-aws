package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	componentArn := d.Get("arn").(string)

	params := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(componentArn),
	}

	resp, err := conn.GetInfrastructureConfiguration(params)

	if err != nil {
		return fmt.Errorf("Error retrieving Component: %s", err)
	}

	return infraconfigDescriptionAttributes(d, resp.InfrastructureConfiguration)
}

func infraconfigDescriptionAttributes(d *schema.ResourceData, component *imagebuilder.InfrastructureConfiguration) error {
	d.SetId(*component.Arn)
	d.Set("datecreated", component.DateCreated)
	d.Set("dateupdated", component.DateUpdated)
	d.Set("date_created", component.DateCreated)
	d.Set("description", component.Description)
	d.Set("instance_profile_name", component.InstanceProfileName)
	d.Set("instance_types", component.InstanceTypes)
	d.Set("key_pair", component.KeyPair)
	d.Set("logging", component.Logging)
	d.Set("name", component.Name)
	d.Set("security_group_ids", component.SecurityGroupIds)
	d.Set("sns_topic_arn", component.SnsTopicArn)
	d.Set("subnet_id", component.SubnetId)
	d.Set("terminate_instance_on_failure", component.TerminateInstanceOnFailure)
	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(component.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
