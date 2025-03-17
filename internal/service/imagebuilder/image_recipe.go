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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_image_recipe", name="Image Recipe")
// @Tags(identifierAttribute="id")
func resourceImageRecipe() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageRecipeCreate,
		ReadWithoutTimeout:   resourceImageRecipeRead,
		UpdateWithoutTimeout: resourceImageRecipeUpdate,
		DeleteWithoutTimeout: resourceImageRecipeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
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
										ValidateFunc: validation.IntBetween(100, 10000),
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
						names.AttrVirtualName: {
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
						names.AttrParameter: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									names.AttrValue: {
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
			names.AttrDescription: {
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
			names.AttrVersion: {
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
	}
}

func resourceImageRecipeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateImageRecipeInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("block_device_mapping"); ok && v.(*schema.Set).Len() > 0 {
		input.BlockDeviceMappings = expandInstanceBlockDeviceMappings(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("component"); ok && len(v.([]any)) > 0 {
		input.Components = expandComponentConfigurations(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_image"); ok {
		input.ParentImage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("systems_manager_agent"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.AdditionalInstanceConfiguration = &awstypes.AdditionalInstanceConfiguration{
			SystemsManagerAgent: expandSystemsManagerAgent(v.([]any)[0].(map[string]any)),
		}
	}

	if v, ok := d.GetOk("user_data_base64"); ok {
		if input.AdditionalInstanceConfiguration == nil {
			input.AdditionalInstanceConfiguration = &awstypes.AdditionalInstanceConfiguration{}
		}
		input.AdditionalInstanceConfiguration.UserDataOverride = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("working_directory"); ok {
		input.WorkingDirectory = aws.String(v.(string))
	}

	output, err := conn.CreateImageRecipe(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image Recipe: %s", err)
	}

	d.SetId(aws.ToString(output.ImageRecipeArn))

	return append(diags, resourceImageRecipeRead(ctx, d, meta)...)
}

func resourceImageRecipeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	imageRecipe, err := findImageRecipeByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Image Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, imageRecipe.Arn)
	if err := d.Set("block_device_mapping", flattenInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting block_device_mapping: %s", err)
	}
	if err := d.Set("component", flattenComponentConfigurations(imageRecipe.Components)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting component: %s", err)
	}
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set(names.AttrDescription, imageRecipe.Description)
	d.Set(names.AttrName, imageRecipe.Name)
	d.Set(names.AttrOwner, imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	if imageRecipe.AdditionalInstanceConfiguration != nil {
		if err := d.Set("systems_manager_agent", []any{flattenSystemsManagerAgent(imageRecipe.AdditionalInstanceConfiguration.SystemsManagerAgent)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting systems_manager_agent: %s", err)
		}
		d.Set("user_data_base64", imageRecipe.AdditionalInstanceConfiguration.UserDataOverride)
	}
	d.Set(names.AttrVersion, imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	setTagsOut(ctx, imageRecipe.Tags)

	return diags
}

func resourceImageRecipeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceImageRecipeRead(ctx, d, meta)...)
}

func resourceImageRecipeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	log.Printf("[DEBUG] Deleting Image Builder Image Recipe: %s", d.Id())
	_, err := conn.DeleteImageRecipe(ctx, &imagebuilder.DeleteImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Image Recipe (%s): %s", d.Id(), err)
	}

	return diags
}

func findImageRecipeByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.ImageRecipe, error) {
	input := &imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(arn),
	}

	output, err := conn.GetImageRecipe(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ImageRecipe == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ImageRecipe, nil
}

func expandComponentConfiguration(tfMap map[string]any) *awstypes.ComponentConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ComponentConfiguration{}

	if v, ok := tfMap["component_arn"].(string); ok && v != "" {
		apiObject.ComponentArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrParameter].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Parameters = expandComponentParameters(v.List())
	}

	return apiObject
}

func expandComponentParameters(tfList []any) []awstypes.ComponentParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ComponentParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

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

func expandComponentParameter(tfMap map[string]any) *awstypes.ComponentParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ComponentParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		apiObject.Value = []string{v}
	}

	return apiObject
}

func expandComponentConfigurations(tfList []any) []awstypes.ComponentConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ComponentConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

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

func expandEBSInstanceBlockDeviceSpecification(tfMap map[string]any) *awstypes.EbsInstanceBlockDeviceSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EbsInstanceBlockDeviceSpecification{}

	if v, null, _ := nullable.Bool(tfMap[names.AttrDeleteOnTermination].(string)).ValueBool(); !null {
		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, null, _ := nullable.Bool(tfMap[names.AttrEncrypted].(string)).ValueBool(); !null {
		apiObject.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Iops = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.VolumeType = awstypes.EbsVolumeType(v)
	}

	return apiObject
}

func expandInstanceBlockDeviceMapping(tfMap map[string]any) awstypes.InstanceBlockDeviceMapping {
	apiObject := awstypes.InstanceBlockDeviceMapping{}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["ebs"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Ebs = expandEBSInstanceBlockDeviceSpecification(v[0].(map[string]any))
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.String("")
	}

	if v, ok := tfMap[names.AttrVirtualName].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandInstanceBlockDeviceMappings(tfList []any) []awstypes.InstanceBlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InstanceBlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandInstanceBlockDeviceMapping(tfMap))
	}

	return apiObjects
}

func expandSystemsManagerAgent(tfMap map[string]any) *awstypes.SystemsManagerAgent {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SystemsManagerAgent{}

	if v, ok := tfMap["uninstall_after_build"].(bool); ok {
		apiObject.UninstallAfterBuild = aws.Bool(v)
	}

	return apiObject
}

func flattenComponentConfiguration(apiObject awstypes.ComponentConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.ComponentArn; v != nil {
		tfMap["component_arn"] = aws.ToString(v)
	}

	if v := apiObject.Parameters; v != nil {
		tfMap[names.AttrParameter] = flattenComponentParameters(v)
	}

	return tfMap
}

func flattenComponentParameters(apiObjects []awstypes.ComponentParameter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenComponentParameter(apiObject))
	}

	return tfList
}

func flattenComponentParameter(apiObject awstypes.ComponentParameter) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		tfMap[names.AttrValue] = aws.StringSlice(v)[0]
	}

	return tfMap
}

func flattenComponentConfigurations(apiObjects []awstypes.ComponentConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenComponentConfiguration(apiObject))
	}

	return tfList
}

func flattenEBSInstanceBlockDeviceSpecification(apiObject *awstypes.EbsInstanceBlockDeviceSpecification) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap[names.AttrDeleteOnTermination] = strconv.FormatBool(aws.ToBool(v))
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap[names.AttrEncrypted] = strconv.FormatBool(aws.ToBool(v))
	}

	if v := apiObject.Iops; v != nil {
		tfMap[names.AttrIOPS] = aws.ToInt32(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}

	if v := apiObject.SnapshotId; v != nil {
		tfMap[names.AttrSnapshotID] = aws.ToString(v)
	}

	if v := apiObject.Throughput; v != nil {
		tfMap[names.AttrThroughput] = aws.ToInt32(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap[names.AttrVolumeSize] = aws.ToInt32(v)
	}

	tfMap[names.AttrVolumeType] = apiObject.VolumeType

	return tfMap
}

func flattenInstanceBlockDeviceMapping(apiObject awstypes.InstanceBlockDeviceMapping) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.DeviceName; v != nil {
		tfMap[names.AttrDeviceName] = aws.ToString(v)
	}

	if v := apiObject.Ebs; v != nil {
		tfMap["ebs"] = []any{flattenEBSInstanceBlockDeviceSpecification(v)}
	}

	if v := apiObject.NoDevice; v != nil {
		tfMap["no_device"] = true
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap[names.AttrVirtualName] = aws.ToString(v)
	}

	return tfMap
}

func flattenInstanceBlockDeviceMappings(apiObjects []awstypes.InstanceBlockDeviceMapping) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenInstanceBlockDeviceMapping(apiObject))
	}

	return tfList
}

func flattenSystemsManagerAgent(apiObject *awstypes.SystemsManagerAgent) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.UninstallAfterBuild; v != nil {
		tfMap["uninstall_after_build"] = aws.ToBool(v)
	}

	return tfMap
}
