package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderImagePipeline() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderImagePipelineRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"image_recipe_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"timeout_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type: schema.TypeString,
				Computed: true,
			},
			"schedule": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_execution_start_condition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schedule_expression": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsImageBuilderImagePipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	resp, err := conn.GetImagePipeline(&imagebuilder.GetImagePipelineInput{
		ImagePipelineArn: aws.String(d.Get("arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("error reading Image Pipeline (%s): %s", d.Id(), err)
	}

	d.SetId(*resp.ImagePipeline.Arn)
	d.Set("description", resp.ImagePipeline.Description)
	d.Set("distribution_configuration_arn", resp.ImagePipeline.DistributionConfigurationArn)
	d.Set("enhanced_image_metadata_enabled", resp.ImagePipeline.EnhancedImageMetadataEnabled)
	d.Set("image_recipe_arn", resp.ImagePipeline.ImageRecipeArn)
	d.Set("image_tests_configuration", flattenAwsImageBuilderTestConfiguration(resp.ImagePipeline.ImageTestsConfiguration))
	d.Set("infrastructure_configuration_arn", resp.ImagePipeline.InfrastructureConfigurationArn)
	d.Set("name", resp.ImagePipeline.Name)
	d.Set("platform", resp.ImagePipeline.Platform)
	if resp.ImagePipeline.Schedule != nil {
		d.Set("schedule", flattenAwsImageBuilderPipelineSchedule(resp.ImagePipeline.Schedule))
	}
	d.Set("status", resp.ImagePipeline.Status)

	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(resp.ImagePipeline.Tags).IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
