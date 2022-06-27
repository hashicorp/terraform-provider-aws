package imagebuilder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInfrastructureConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"instance_profile_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"resource_tags": tftags.TagsSchemaComputed(),
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sns_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceInfrastructureConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetInfrastructureConfigurationInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.InfrastructureConfigurationArn = aws.String(v.(string))
	}

	output, err := conn.GetInfrastructureConfiguration(input)

	if err != nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.InfrastructureConfiguration == nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): empty response", d.Id())
	}

	infrastructureConfiguration := output.InfrastructureConfiguration

	d.SetId(aws.StringValue(infrastructureConfiguration.Arn))
	d.Set("arn", infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set("description", infrastructureConfiguration.Description)

	if infrastructureConfiguration.InstanceMetadataOptions != nil {
		d.Set("instance_metadata_options", []interface{}{flattenInstanceMetadataOptions(infrastructureConfiguration.InstanceMetadataOptions)})
	} else {
		d.Set("instance_metadata_options", nil)
	}

	d.Set("instance_profile_name", infrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", aws.StringValueSlice(infrastructureConfiguration.InstanceTypes))
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		d.Set("logging", []interface{}{flattenLogging(infrastructureConfiguration.Logging)})
	} else {
		d.Set("logging", nil)
	}
	d.Set("name", infrastructureConfiguration.Name)
	d.Set("resource_tags", KeyValueTags(infrastructureConfiguration.ResourceTags).Map())
	d.Set("security_group_ids", aws.StringValueSlice(infrastructureConfiguration.SecurityGroupIds))
	d.Set("sns_topic_arn", infrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", infrastructureConfiguration.SubnetId)
	d.Set("tags", KeyValueTags(infrastructureConfiguration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	return nil
}
