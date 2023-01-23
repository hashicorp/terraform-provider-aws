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

func DataSourceContainerRecipe() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContainerRecipeRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
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
			"container_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dockerfile_template_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"image": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"kms_key_id": {
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
			"target_repository": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func dataSourceContainerRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetContainerRecipeInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ContainerRecipeArn = aws.String(v.(string))
	}

	output, err := conn.GetContainerRecipeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Container Recipe (%s): %s", aws.StringValue(input.ContainerRecipeArn), err)
	}

	if output == nil || output.ContainerRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Container Recipe (%s): empty response", aws.StringValue(input.ContainerRecipeArn))
	}

	containerRecipe := output.ContainerRecipe

	d.SetId(aws.StringValue(containerRecipe.Arn))
	d.Set("arn", containerRecipe.Arn)
	d.Set("component", flattenComponentConfigurations(containerRecipe.Components))
	d.Set("container_type", containerRecipe.ContainerType)
	d.Set("date_created", containerRecipe.DateCreated)
	d.Set("description", containerRecipe.Description)
	d.Set("dockerfile_template_data", containerRecipe.DockerfileTemplateData)
	d.Set("encrypted", containerRecipe.Encrypted)

	if containerRecipe.InstanceConfiguration != nil {
		d.Set("instance_configuration", []interface{}{flattenInstanceConfiguration(containerRecipe.InstanceConfiguration)})
	} else {
		d.Set("instance_configuration", nil)
	}

	d.Set("kms_key_id", containerRecipe.KmsKeyId)
	d.Set("name", containerRecipe.Name)
	d.Set("owner", containerRecipe.Owner)
	d.Set("parent_image", containerRecipe.ParentImage)
	d.Set("platform", containerRecipe.Platform)
	d.Set("tags", KeyValueTags(containerRecipe.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("target_repository", []interface{}{flattenTargetContainerRepository(containerRecipe.TargetRepository)})
	d.Set("version", containerRecipe.Version)
	d.Set("working_directory", containerRecipe.WorkingDirectory)

	return diags
}
