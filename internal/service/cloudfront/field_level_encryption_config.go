// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_field_level_encryption_config", name="Field-level Encryption Config")
func resourceFieldLevelEncryptionConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFieldLevelEncryptionConfigCreate,
		ReadWithoutTimeout:   resourceFieldLevelEncryptionConfigRead,
		UpdateWithoutTimeout: resourceFieldLevelEncryptionConfigUpdate,
		DeleteWithoutTimeout: resourceFieldLevelEncryptionConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_type_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_type_profiles": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrContentType: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrFormat: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.Format](),
												},
												"profile_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"forward_when_content_type_is_unknown": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_arg_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"forward_when_query_arg_profile_is_unknown": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"query_arg_profiles": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"profile_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"query_arg": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFieldLevelEncryptionConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	apiObject := &awstypes.FieldLevelEncryptionConfig{
		CallerReference: aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ContentTypeProfileConfig = expandContentTypeProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("query_arg_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.QueryArgProfileConfig = expandQueryArgProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.CreateFieldLevelEncryptionConfigInput{
		FieldLevelEncryptionConfig: apiObject,
	}

	output, err := conn.CreateFieldLevelEncryptionConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Field-level Encryption Config: %s", err)
	}

	d.SetId(aws.ToString(output.FieldLevelEncryption.Id))

	return append(diags, resourceFieldLevelEncryptionConfigRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findFieldLevelEncryptionConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Field-level Encryption Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	apiObject := output.FieldLevelEncryptionConfig
	d.Set("caller_reference", apiObject.CallerReference)
	d.Set(names.AttrComment, apiObject.Comment)
	if apiObject.ContentTypeProfileConfig != nil {
		if err := d.Set("content_type_profile_config", []interface{}{flattenContentTypeProfileConfig(apiObject.ContentTypeProfileConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting content_type_profile_config: %s", err)
		}
	} else {
		d.Set("content_type_profile_config", nil)
	}
	d.Set("etag", output.ETag)
	if apiObject.QueryArgProfileConfig != nil {
		if err := d.Set("query_arg_profile_config", []interface{}{flattenQueryArgProfileConfig(apiObject.QueryArgProfileConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting query_arg_profile_config: %s", err)
		}
	} else {
		d.Set("query_arg_profile_config", nil)
	}

	return diags
}

func resourceFieldLevelEncryptionConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	apiObject := &awstypes.FieldLevelEncryptionConfig{
		CallerReference: aws.String(d.Get("caller_reference").(string)),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ContentTypeProfileConfig = expandContentTypeProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("query_arg_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.QueryArgProfileConfig = expandQueryArgProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.UpdateFieldLevelEncryptionConfigInput{
		FieldLevelEncryptionConfig: apiObject,
		Id:                         aws.String(d.Id()),
		IfMatch:                    aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateFieldLevelEncryptionConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFieldLevelEncryptionConfigRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Field-level Encryption Config: (%s)", d.Id())
	_, err := conn.DeleteFieldLevelEncryptionConfig(ctx, &cloudfront.DeleteFieldLevelEncryptionConfigInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchFieldLevelEncryptionConfig](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findFieldLevelEncryptionConfigByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetFieldLevelEncryptionConfigOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchFieldLevelEncryptionConfig](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FieldLevelEncryptionConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandContentTypeProfileConfig(tfMap map[string]interface{}) *awstypes.ContentTypeProfileConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContentTypeProfileConfig{}

	if v, ok := tfMap["content_type_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContentTypeProfiles = expandContentTypeProfiles(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["forward_when_content_type_is_unknown"].(bool); ok {
		apiObject.ForwardWhenContentTypeIsUnknown = aws.Bool(v)
	}

	return apiObject
}

func expandContentTypeProfiles(tfMap map[string]interface{}) *awstypes.ContentTypeProfiles {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContentTypeProfiles{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandContentTypeProfileItems(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandContentTypeProfile(tfMap map[string]interface{}) *awstypes.ContentTypeProfile {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContentTypeProfile{}

	if v, ok := tfMap[names.AttrContentType].(string); ok && v != "" {
		apiObject.ContentType = aws.String(v)
	}

	if v, ok := tfMap[names.AttrFormat].(string); ok && v != "" {
		apiObject.Format = awstypes.Format(v)
	}

	if v, ok := tfMap["profile_id"].(string); ok && v != "" {
		apiObject.ProfileId = aws.String(v)
	}

	return apiObject
}

func expandContentTypeProfileItems(tfList []interface{}) []awstypes.ContentTypeProfile {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ContentTypeProfile

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandContentTypeProfile(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandQueryArgProfileConfig(tfMap map[string]interface{}) *awstypes.QueryArgProfileConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.QueryArgProfileConfig{}

	if v, ok := tfMap["forward_when_query_arg_profile_is_unknown"].(bool); ok {
		apiObject.ForwardWhenQueryArgProfileIsUnknown = aws.Bool(v)
	}

	if v, ok := tfMap["query_arg_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.QueryArgProfiles = expandQueryArgProfiles(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandQueryArgProfiles(tfMap map[string]interface{}) *awstypes.QueryArgProfiles {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.QueryArgProfiles{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandQueryArgProfileItems(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandQueryArgProfile(tfMap map[string]interface{}) *awstypes.QueryArgProfile {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.QueryArgProfile{}

	if v, ok := tfMap["profile_id"].(string); ok && v != "" {
		apiObject.ProfileId = aws.String(v)
	}

	if v, ok := tfMap["query_arg"].(string); ok && v != "" {
		apiObject.QueryArg = aws.String(v)
	}

	return apiObject
}

func expandQueryArgProfileItems(tfList []interface{}) []awstypes.QueryArgProfile {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.QueryArgProfile

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandQueryArgProfile(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenContentTypeProfileConfig(apiObject *awstypes.ContentTypeProfileConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenContentTypeProfiles(apiObject.ContentTypeProfiles); len(v) > 0 {
		tfMap["content_type_profiles"] = []interface{}{v}
	}

	if v := apiObject.ForwardWhenContentTypeIsUnknown; v != nil {
		tfMap["forward_when_content_type_is_unknown"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenContentTypeProfiles(apiObject *awstypes.ContentTypeProfiles) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenContentTypeProfileItems(v)
	}

	return tfMap
}

func flattenContentTypeProfile(apiObject *awstypes.ContentTypeProfile) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrFormat: apiObject.Format,
	}

	if v := apiObject.ContentType; v != nil {
		tfMap[names.AttrContentType] = aws.ToString(v)
	}

	if v := apiObject.ProfileId; v != nil {
		tfMap["profile_id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenContentTypeProfileItems(apiObjects []awstypes.ContentTypeProfile) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if v := flattenContentTypeProfile(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenQueryArgProfileConfig(apiObject *awstypes.QueryArgProfileConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ForwardWhenQueryArgProfileIsUnknown; v != nil {
		tfMap["forward_when_query_arg_profile_is_unknown"] = aws.ToBool(v)
	}

	if v := flattenQueryArgProfiles(apiObject.QueryArgProfiles); len(v) > 0 {
		tfMap["query_arg_profiles"] = []interface{}{v}
	}

	return tfMap
}

func flattenQueryArgProfiles(apiObject *awstypes.QueryArgProfiles) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenQueryArgProfileItems(v)
	}

	return tfMap
}

func flattenQueryArgProfile(apiObject *awstypes.QueryArgProfile) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ProfileId; v != nil {
		tfMap["profile_id"] = aws.ToString(v)
	}

	if v := apiObject.QueryArg; v != nil {
		tfMap["query_arg"] = aws.ToString(v)
	}

	return tfMap
}

func flattenQueryArgProfileItems(apiObjects []awstypes.QueryArgProfile) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if v := flattenQueryArgProfile(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}
