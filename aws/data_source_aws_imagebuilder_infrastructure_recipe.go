package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

	resp, err := conn.GetImageRecipe(&imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading Recipe (%s): %s", d.Id(), err)
	}

	return recipeDescriptionAttributes(d, resp.ImageRecipe)
}

func recipeDescriptionAttributes(d *schema.ResourceData, imageRecipe *imagebuilder.ImageRecipe) error {
	d.SetId(*imageRecipe.Arn)
	d.Set("arn", imageRecipe.Arn)
	d.Set("block_device_mappings", imageRecipe.BlockDeviceMappings)
	d.Set("components", imageRecipe.Components)
	d.Set("datecreated", imageRecipe.DateCreated)
	d.Set("description", imageRecipe.Description)
	d.Set("name", imageRecipe.Name)
	d.Set("owner", imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	d.Set("semantic_version", imageRecipe.Version)

	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(imageRecipe.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
