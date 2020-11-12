package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderRecipe() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderRecipeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"ebs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
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
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"components": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"datecreated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsImageBuilderRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetImageRecipe(&imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Get("arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("error reading Recipe (%s): %s", d.Id(), err)
	}

	d.SetId(*resp.ImageRecipe.Arn)
	d.Set("block_device_mappings", flattenImageBuilderRecipeBlockDeviceMappings(resp.ImageRecipe.BlockDeviceMappings))
	d.Set("components", flattenImageBuilderRecipeComponents(resp.ImageRecipe.Components))
	d.Set("datecreated", resp.ImageRecipe.DateCreated)
	d.Set("description", resp.ImageRecipe.Description)
	d.Set("name", resp.ImageRecipe.Name)
	d.Set("owner", resp.ImageRecipe.Owner)
	d.Set("parent_image", resp.ImageRecipe.ParentImage)
	d.Set("platform", resp.ImageRecipe.Platform)
	d.Set("semantic_version", resp.ImageRecipe.Version)

	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(resp.ImageRecipe.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
