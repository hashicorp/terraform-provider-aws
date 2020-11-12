package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
)

func datasourceAwsImageBuilderDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: datasourceAwsImageBuilderDistributionConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
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
			"distributions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami_distribution_configuration": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ami_tags": tagsSchema(),
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"launch_permission": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"user_groups": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"user_ids": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"license_configuration_arn": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func datasourceAwsImageBuilderDistributionConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	resp, err := conn.GetDistributionConfiguration(&imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Get("arn").(string)),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] DistributionConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DistributionConfiguration (%s): %s", d.Id(), err)
	}

	d.SetId(*resp.DistributionConfiguration.Arn)
	d.Set("date_created", resp.DistributionConfiguration.DateCreated)
	d.Set("date_updated", resp.DistributionConfiguration.DateUpdated)
	d.Set("description", resp.DistributionConfiguration.Description)
	d.Set("distributions", flattenAwsImageBuilderDistributions(resp.DistributionConfiguration.Distributions))
	d.Set("name", resp.DistributionConfiguration.Name)

	if err != nil {
		return fmt.Errorf("error listing tags for DistributionConfiguration (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(resp.DistributionConfiguration.Tags).IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
