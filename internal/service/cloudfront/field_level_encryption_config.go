package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFieldLevelEncryptionConfig() *schema.Resource {
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
			"comment": {
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
												"content_type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"format": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cloudfront.Format_Values(), false),
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
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	apiObject := &cloudfront.FieldLevelEncryptionConfig{
		CallerReference: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("comment"); ok {
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

	log.Printf("[DEBUG] Creating CloudFront Field-level Encryption Config: (%s)", input)
	output, err := conn.CreateFieldLevelEncryptionConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.FieldLevelEncryption.Id))

	return append(diags, resourceFieldLevelEncryptionConfigRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	output, err := FindFieldLevelEncryptionConfigByID(ctx, conn, d.Id())

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
	d.Set("comment", apiObject.Comment)
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
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	apiObject := &cloudfront.FieldLevelEncryptionConfig{
		CallerReference: aws.String(d.Get("caller_reference").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
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

	log.Printf("[DEBUG] Updating CloudFront Field-level Encryption Config: (%s)", input)
	_, err := conn.UpdateFieldLevelEncryptionConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFieldLevelEncryptionConfigRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	log.Printf("[DEBUG] Deleting CloudFront Field-level Encryption Config: (%s)", d.Id())
	_, err := conn.DeleteFieldLevelEncryptionConfigWithContext(ctx, &cloudfront.DeleteFieldLevelEncryptionConfigInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFieldLevelEncryptionConfig) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Field-level Encryption Config (%s): %s", d.Id(), err)
	}

	return diags
}

func expandContentTypeProfileConfig(tfMap map[string]interface{}) *cloudfront.ContentTypeProfileConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ContentTypeProfileConfig{}

	if v, ok := tfMap["content_type_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContentTypeProfiles = expandContentTypeProfiles(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["forward_when_content_type_is_unknown"].(bool); ok {
		apiObject.ForwardWhenContentTypeIsUnknown = aws.Bool(v)
	}

	return apiObject
}

func expandContentTypeProfiles(tfMap map[string]interface{}) *cloudfront.ContentTypeProfiles {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ContentTypeProfiles{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandContentTypeProfileItems(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandContentTypeProfile(tfMap map[string]interface{}) *cloudfront.ContentTypeProfile {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ContentTypeProfile{}

	if v, ok := tfMap["content_type"].(string); ok && v != "" {
		apiObject.ContentType = aws.String(v)
	}

	if v, ok := tfMap["format"].(string); ok && v != "" {
		apiObject.Format = aws.String(v)
	}

	if v, ok := tfMap["profile_id"].(string); ok && v != "" {
		apiObject.ProfileId = aws.String(v)
	}

	return apiObject
}

func expandContentTypeProfileItems(tfList []interface{}) []*cloudfront.ContentTypeProfile {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*cloudfront.ContentTypeProfile

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandContentTypeProfile(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandQueryArgProfileConfig(tfMap map[string]interface{}) *cloudfront.QueryArgProfileConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.QueryArgProfileConfig{}

	if v, ok := tfMap["forward_when_query_arg_profile_is_unknown"].(bool); ok {
		apiObject.ForwardWhenQueryArgProfileIsUnknown = aws.Bool(v)
	}

	if v, ok := tfMap["query_arg_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.QueryArgProfiles = expandQueryArgProfiles(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandQueryArgProfiles(tfMap map[string]interface{}) *cloudfront.QueryArgProfiles {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.QueryArgProfiles{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandQueryArgProfileItems(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandQueryArgProfile(tfMap map[string]interface{}) *cloudfront.QueryArgProfile {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.QueryArgProfile{}

	if v, ok := tfMap["profile_id"].(string); ok && v != "" {
		apiObject.ProfileId = aws.String(v)
	}

	if v, ok := tfMap["query_arg"].(string); ok && v != "" {
		apiObject.QueryArg = aws.String(v)
	}

	return apiObject
}

func expandQueryArgProfileItems(tfList []interface{}) []*cloudfront.QueryArgProfile {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*cloudfront.QueryArgProfile

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandQueryArgProfile(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenContentTypeProfileConfig(apiObject *cloudfront.ContentTypeProfileConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenContentTypeProfiles(apiObject.ContentTypeProfiles); len(v) > 0 {
		tfMap["content_type_profiles"] = []interface{}{v}
	}

	if v := apiObject.ForwardWhenContentTypeIsUnknown; v != nil {
		tfMap["forward_when_content_type_is_unknown"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenContentTypeProfiles(apiObject *cloudfront.ContentTypeProfiles) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenContentTypeProfileItems(v)
	}

	return tfMap
}

func flattenContentTypeProfile(apiObject *cloudfront.ContentTypeProfile) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContentType; v != nil {
		tfMap["content_type"] = aws.StringValue(v)
	}

	if v := apiObject.Format; v != nil {
		tfMap["format"] = aws.StringValue(v)
	}

	if v := apiObject.ProfileId; v != nil {
		tfMap["profile_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenContentTypeProfileItems(apiObjects []*cloudfront.ContentTypeProfile) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenContentTypeProfile(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenQueryArgProfileConfig(apiObject *cloudfront.QueryArgProfileConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ForwardWhenQueryArgProfileIsUnknown; v != nil {
		tfMap["forward_when_query_arg_profile_is_unknown"] = aws.BoolValue(v)
	}

	if v := flattenQueryArgProfiles(apiObject.QueryArgProfiles); len(v) > 0 {
		tfMap["query_arg_profiles"] = []interface{}{v}
	}

	return tfMap
}

func flattenQueryArgProfiles(apiObject *cloudfront.QueryArgProfiles) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenQueryArgProfileItems(v)
	}

	return tfMap
}

func flattenQueryArgProfile(apiObject *cloudfront.QueryArgProfile) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ProfileId; v != nil {
		tfMap["profile_id"] = aws.StringValue(v)
	}

	if v := apiObject.QueryArg; v != nil {
		tfMap["query_arg"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenQueryArgProfileItems(apiObjects []*cloudfront.QueryArgProfile) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenQueryArgProfile(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}
