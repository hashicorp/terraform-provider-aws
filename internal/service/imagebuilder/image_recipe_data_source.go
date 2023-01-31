package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceImageRecipe() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRecipeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"block_device_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
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
									"throughput": {
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
			"component": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"date_created": {
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
			"tags": tftags.TagsSchema(),
			"user_data_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"working_directory": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetImageRecipeInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	output, err := conn.GetImageRecipeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): %s", aws.StringValue(input.ImageRecipeArn), err)
	}

	if output == nil || output.ImageRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): empty response", aws.StringValue(input.ImageRecipeArn))
	}

	imageRecipe := output.ImageRecipe

	d.SetId(aws.StringValue(imageRecipe.Arn))
	d.Set("arn", imageRecipe.Arn)
	d.Set("block_device_mapping", flattenInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings))
	d.Set("component", flattenComponentConfigurations(imageRecipe.Components))
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set("description", imageRecipe.Description)
	d.Set("name", imageRecipe.Name)
	d.Set("owner", imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	d.Set("tags", KeyValueTags(imageRecipe.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	if imageRecipe.AdditionalInstanceConfiguration != nil {
		d.Set("user_data_base64", imageRecipe.AdditionalInstanceConfiguration.UserDataOverride)
	}

	d.Set("version", imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	return diags
}
