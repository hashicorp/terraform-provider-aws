// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_image_recipe", name="Image Recipe")
// @Tags(identifierAttribute="id")
func ResourceImageRecipe() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageRecipeCreate,
		ReadWithoutTimeout:   resourceImageRecipeRead,
		UpdateWithoutTimeout: resourceImageRecipeUpdate,
		DeleteWithoutTimeout: resourceImageRecipeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
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
									"delete_on_termination": {
										Type:             nullable.TypeNullableBool,
										Optional:         true,
										ForceNew:         true,
										DiffSuppressFunc: nullable.DiffSuppressNullableBool,
										ValidateFunc:     nullable.ValidateTypeStringNullableBool,
									},
									"encrypted": {
										Type:             nullable.TypeNullableBool,
										Optional:         true,
										ForceNew:         true,
										DiffSuppressFunc: nullable.DiffSuppressNullableBool,
										ValidateFunc:     nullable.ValidateTypeStringNullableBool,
									},
									"iops": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(100, 10000),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"snapshot_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"throughput": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(125, 1000),
									},
									"volume_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 16000),
									},
									"volume_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EbsVolumeType](),
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
							Computed: true,
							ForceNew: true,
						},
						"virtual_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
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
						"parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"owner": {
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
			"systems_manager_agent": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uninstall_after_build": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_data_base64": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 21847),
					verify.ValidBase64String,
				),
			},
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
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

func resourceImageRecipeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateImageRecipeInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("block_device_mapping"); ok && v.(*schema.Set).Len() > 0 {
		input.BlockDeviceMappings = expandInstanceBlockDeviceMappings(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("component"); ok && len(v.([]interface{})) > 0 {
		input.Components = expandComponentConfigurations(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_image"); ok {
		input.ParentImage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("systems_manager_agent"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AdditionalInstanceConfiguration = &awstypes.AdditionalInstanceConfiguration{
			SystemsManagerAgent: expandSystemsManagerAgent(v.([]interface{})[0].(map[string]interface{})),
		}
	}

	if v, ok := d.GetOk("user_data_base64"); ok {
		if input.AdditionalInstanceConfiguration == nil {
			input.AdditionalInstanceConfiguration = &awstypes.AdditionalInstanceConfiguration{}
		}
		input.AdditionalInstanceConfiguration.UserDataOverride = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version"); ok {
		input.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("working_directory"); ok {
		input.WorkingDirectory = aws.String(v.(string))
	}

	output, err := conn.CreateImageRecipe(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image Recipe: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image Recipe: empty response")
	}

	d.SetId(aws.ToString(output.ImageRecipeArn))

	return append(diags, resourceImageRecipeRead(ctx, d, meta)...)
}

func resourceImageRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	}

	output, err := conn.GetImageRecipe(ctx, input)

	if !d.IsNewResource() && errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		log.Printf("[WARN] Image Builder Image Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image Recipe (%s): %s", d.Id(), err)
	}

	if output == nil || output.ImageRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image Recipe (%s): empty response", d.Id())
	}

	imageRecipe := output.ImageRecipe

	d.Set("arn", imageRecipe.Arn)
	d.Set("block_device_mapping", flattenInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings))
	d.Set("component", flattenComponentConfigurations(imageRecipe.Components))
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set("description", imageRecipe.Description)
	d.Set("name", imageRecipe.Name)
	d.Set("owner", imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)

	setTagsOut(ctx, imageRecipe.Tags)

	if imageRecipe.AdditionalInstanceConfiguration != nil {
		d.Set("systems_manager_agent", []interface{}{flattenSystemsManagerAgent(imageRecipe.AdditionalInstanceConfiguration.SystemsManagerAgent)})
		d.Set("user_data_base64", imageRecipe.AdditionalInstanceConfiguration.UserDataOverride)
	}

	d.Set("version", imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	return diags
}

func resourceImageRecipeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceImageRecipeRead(ctx, d, meta)...)
}

func resourceImageRecipeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.DeleteImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteImageRecipe(ctx, input)

	if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Image Recipe (%s): %s", d.Id(), err)
	}

	return diags
}

func expandComponentConfiguration(tfMap map[string]interface{}) *awstypes.ComponentConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ComponentConfiguration{}

	if v, ok := tfMap["component_arn"].(string); ok && v != "" {
		apiObject.ComponentArn = aws.String(v)
	}

	if v, ok := tfMap["parameter"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Parameters = expandComponentParameters(v.List())
	}

	return apiObject
}

func expandComponentParameters(tfList []interface{}) []awstypes.ComponentParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ComponentParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandComponentParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandComponentParameter(tfMap map[string]interface{}) *awstypes.ComponentParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ComponentParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		apiObject.Value = []string{v}
	}

	return apiObject
}

func expandComponentConfigurations(tfList []interface{}) []awstypes.ComponentConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ComponentConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandComponentConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandEBSInstanceBlockDeviceSpecification(tfMap map[string]interface{}) *awstypes.EbsInstanceBlockDeviceSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EbsInstanceBlockDeviceSpecification{}

	if v, null, _ := nullable.Bool(tfMap["delete_on_termination"].(string)).Value(); !null {
		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, null, _ := nullable.Bool(tfMap["encrypted"].(string)).Value(); !null {
		apiObject.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Iops = aws.Int32(int32(v))
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.VolumeType = awstypes.EbsVolumeType(v)
	}

	return apiObject
}

func expandInstanceBlockDeviceMapping(tfMap map[string]interface{}) awstypes.InstanceBlockDeviceMapping {
	apiObject := awstypes.InstanceBlockDeviceMapping{}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["ebs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Ebs = expandEBSInstanceBlockDeviceSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.String("")
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandInstanceBlockDeviceMappings(tfList []interface{}) []awstypes.InstanceBlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InstanceBlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandInstanceBlockDeviceMapping(tfMap))
	}

	return apiObjects
}

func expandSystemsManagerAgent(tfMap map[string]interface{}) *awstypes.SystemsManagerAgent {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SystemsManagerAgent{}

	if v, ok := tfMap["uninstall_after_build"].(bool); ok {
		apiObject.UninstallAfterBuild = aws.Bool(v)
	}

	return apiObject
}

func flattenComponentConfiguration(apiObject awstypes.ComponentConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.ComponentArn; v != nil {
		tfMap["component_arn"] = aws.ToString(v)
	}

	if v := apiObject.Parameters; v != nil {
		tfMap["parameter"] = flattenComponentParameters(v)
	}

	return tfMap
}

func flattenComponentParameters(apiObjects []awstypes.ComponentParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenComponentParameter(apiObject))
	}

	return tfList
}

func flattenComponentParameter(apiObject awstypes.ComponentParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		tfMap["value"] = aws.StringSlice(v)[0]
	}

	return tfMap
}

func flattenComponentConfigurations(apiObjects []awstypes.ComponentConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenComponentConfiguration(apiObject))
	}

	return tfList
}

func flattenEBSInstanceBlockDeviceSpecification(apiObject *awstypes.EbsInstanceBlockDeviceSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = strconv.FormatBool(aws.ToBool(v))
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap["encrypted"] = strconv.FormatBool(aws.ToBool(v))
	}

	if v := apiObject.Iops; v != nil {
		tfMap["iops"] = aws.ToInt32(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.ToString(v)
	}

	if v := apiObject.SnapshotId; v != nil {
		tfMap["snapshot_id"] = aws.ToString(v)
	}

	if v := apiObject.Throughput; v != nil {
		tfMap["throughput"] = aws.ToInt32(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap["volume_size"] = aws.ToInt32(v)
	}

	tfMap["volume_type"] = apiObject.VolumeType

	return tfMap
}

func flattenInstanceBlockDeviceMapping(apiObject awstypes.InstanceBlockDeviceMapping) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.ToString(v)
	}

	if v := apiObject.Ebs; v != nil {
		tfMap["ebs"] = []interface{}{flattenEBSInstanceBlockDeviceSpecification(v)}
	}

	if v := apiObject.NoDevice; v != nil {
		tfMap["no_device"] = true
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap["virtual_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenInstanceBlockDeviceMappings(apiObjects []awstypes.InstanceBlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenInstanceBlockDeviceMapping(apiObject))
	}

	return tfList
}

func flattenSystemsManagerAgent(apiObject *awstypes.SystemsManagerAgent) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.UninstallAfterBuild; v != nil {
		tfMap["uninstall_after_build"] = aws.ToBool(v)
	}

	return tfMap
}
