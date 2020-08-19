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
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"ebs": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"iops": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(100, 10000),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"volume_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 16000),
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											imagebuilder.EbsVolumeTypeStandard,
											imagebuilder.EbsVolumeTypeIo1,
											imagebuilder.EbsVolumeTypeGp2,
											imagebuilder.EbsVolumeTypeSc1,
											imagebuilder.EbsVolumeTypeSt1,
										}, true),
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"virtual_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"components": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"datecreated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"semantic_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsImageBuilderRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetImageRecipe(&imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading Recipe (%s): %s", d.Id(), err)
	}

	d.SetId(*resp.ImageRecipe.Arn)
	d.Set("arn", resp.ImageRecipe.Arn)
	d.Set("block_device_mappings", resp.ImageRecipe.BlockDeviceMappings)
	d.Set("components", resp.ImageRecipe.Components)
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
