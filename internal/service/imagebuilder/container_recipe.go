// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_container_recipe", name="Container Recipe")
// @Tags(identifierAttribute="id")
func ResourceContainerRecipe() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerRecipeCreate,
		ReadWithoutTimeout:   resourceContainerRecipeRead,
		UpdateWithoutTimeout: resourceContainerRecipeUpdate,
		DeleteWithoutTimeout: resourceContainerRecipeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"component": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrParameter: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"container_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"DOCKER"}, false),
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"dockerfile_template_data": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dockerfile_template_data", "dockerfile_template_uri"},
				ValidateFunc: validation.StringLenBetween(1, 16000),
			},
			"dockerfile_template_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dockerfile_template_data", "dockerfile_template_uri"},
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_device_mapping": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDeviceName: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"ebs": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDeleteOnTermination: {
													Type:             nullable.TypeNullableBool,
													Optional:         true,
													ForceNew:         true,
													DiffSuppressFunc: nullable.DiffSuppressNullableBool,
													ValidateFunc:     nullable.ValidateTypeStringNullableBool,
												},
												names.AttrEncrypted: {
													Type:             nullable.TypeNullableBool,
													Optional:         true,
													ForceNew:         true,
													DiffSuppressFunc: nullable.DiffSuppressNullableBool,
													ValidateFunc:     nullable.ValidateTypeStringNullableBool,
												},
												names.AttrIOPS: {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(100, 64000),
												},
												names.AttrKMSKeyID: {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
												names.AttrSnapshotID: {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
												names.AttrThroughput: {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(125, 1000),
												},
												names.AttrVolumeSize: {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntBetween(1, 16000),
												},
												names.AttrVolumeType: {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(imagebuilder.EbsVolumeType_Values(), false),
												},
											},
										},
									},
									"no_device": {
										// Use TypeBool to allow an "unspecified" value of false,
										// since the API uses an empty string ("") as true and
										// this is not compatible with TypeString's zero value.
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
									names.AttrVirtualName: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_override": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Linux", "Windows"}, false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_repository": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRepositoryName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"service": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"ECR"}, false),
						},
					},
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"working_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceContainerRecipeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.CreateContainerRecipeInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("component"); ok && len(v.([]interface{})) > 0 {
		input.Components = expandComponentConfigurations(v.([]interface{}))
	}

	if v, ok := d.GetOk("container_type"); ok {
		input.ContainerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dockerfile_template_data"); ok {
		input.DockerfileTemplateData = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dockerfile_template_uri"); ok {
		input.DockerfileTemplateUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceConfiguration = expandInstanceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_image"); ok {
		input.ParentImage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform_override"); ok {
		input.PlatformOverride = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_repository"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TargetRepository = expandTargetContainerRepository(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.SemanticVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("working_directory"); ok {
		input.WorkingDirectory = aws.String(v.(string))
	}

	output, err := conn.CreateContainerRecipeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Container Recipe: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Container Recipe: empty response")
	}

	d.SetId(aws.StringValue(output.ContainerRecipeArn))

	return append(diags, resourceContainerRecipeRead(ctx, d, meta)...)
}

func resourceContainerRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.GetContainerRecipeInput{
		ContainerRecipeArn: aws.String(d.Id()),
	}

	output, err := conn.GetContainerRecipeWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Container Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Container Recipe (%s): %s", d.Id(), err)
	}

	if output == nil || output.ContainerRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Container Recipe (%s): empty response", d.Id())
	}

	containerRecipe := output.ContainerRecipe

	d.Set(names.AttrARN, containerRecipe.Arn)
	d.Set("component", flattenComponentConfigurations(containerRecipe.Components))
	d.Set("container_type", containerRecipe.ContainerType)
	d.Set("date_created", containerRecipe.DateCreated)
	d.Set(names.AttrDescription, containerRecipe.Description)
	d.Set("dockerfile_template_data", containerRecipe.DockerfileTemplateData)
	d.Set(names.AttrEncrypted, containerRecipe.Encrypted)

	if containerRecipe.InstanceConfiguration != nil {
		d.Set("instance_configuration", []interface{}{flattenInstanceConfiguration(containerRecipe.InstanceConfiguration)})
	} else {
		d.Set("instance_configuration", nil)
	}

	d.Set(names.AttrKMSKeyID, containerRecipe.KmsKeyId)
	d.Set(names.AttrName, containerRecipe.Name)
	d.Set(names.AttrOwner, containerRecipe.Owner)
	d.Set("parent_image", containerRecipe.ParentImage)
	d.Set("platform", containerRecipe.Platform)

	setTagsOut(ctx, containerRecipe.Tags)

	d.Set("target_repository", []interface{}{flattenTargetContainerRepository(containerRecipe.TargetRepository)})
	d.Set(names.AttrVersion, containerRecipe.Version)
	d.Set("working_directory", containerRecipe.WorkingDirectory)

	return diags
}

func resourceContainerRecipeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceContainerRecipeRead(ctx, d, meta)...)
}

func resourceContainerRecipeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.DeleteContainerRecipeInput{
		ContainerRecipeArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerRecipeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Container Recipe (%s): %s", d.Id(), err)
	}

	return diags
}

func expandInstanceConfiguration(tfMap map[string]interface{}) *imagebuilder.InstanceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.InstanceConfiguration{}

	if v, ok := tfMap["block_device_mapping"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.BlockDeviceMappings = expandInstanceBlockDeviceMappings(v.List())
	}

	if v, ok := tfMap["image"].(string); ok && v != "" {
		apiObject.Image = aws.String(v)
	}

	return apiObject
}

func flattenInstanceConfiguration(apiObject *imagebuilder.InstanceConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockDeviceMappings; v != nil {
		tfMap["block_device_mapping"] = flattenInstanceBlockDeviceMappings(v)
	}

	if v := apiObject.Image; v != nil {
		tfMap["image"] = aws.StringValue(v)
	}

	return tfMap
}
