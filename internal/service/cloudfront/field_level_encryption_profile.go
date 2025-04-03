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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_field_level_encryption_profile", name="Field-level Encryption Profile")
func resourceFieldLevelEncryptionProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFieldLevelEncryptionProfileCreate,
		ReadWithoutTimeout:   resourceFieldLevelEncryptionProfileRead,
		UpdateWithoutTimeout: resourceFieldLevelEncryptionProfileUpdate,
		DeleteWithoutTimeout: resourceFieldLevelEncryptionProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_entities": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field_patterns": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"provider_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"public_key_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceFieldLevelEncryptionProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	apiObject := &awstypes.FieldLevelEncryptionProfileConfig{
		CallerReference: aws.String(id.UniqueId()),
		Name:            aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_entities"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		apiObject.EncryptionEntities = expandEncryptionEntities(v.([]any)[0].(map[string]any))
	}

	input := &cloudfront.CreateFieldLevelEncryptionProfileInput{
		FieldLevelEncryptionProfileConfig: apiObject,
	}

	output, err := conn.CreateFieldLevelEncryptionProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Field-level Encryption Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FieldLevelEncryptionProfile.Id))

	return append(diags, resourceFieldLevelEncryptionProfileRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findFieldLevelEncryptionProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Field-level Encryption Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Field-level Encryption Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fieldLevelEncryptionProfileARN(ctx, meta.(*conns.AWSClient), d.Id()))
	apiObject := output.FieldLevelEncryptionProfile.FieldLevelEncryptionProfileConfig
	d.Set("caller_reference", apiObject.CallerReference)
	d.Set(names.AttrComment, apiObject.Comment)
	if apiObject.EncryptionEntities != nil {
		if err := d.Set("encryption_entities", []any{flattenEncryptionEntities(apiObject.EncryptionEntities)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting encryption_entities: %s", err)
		}
	} else {
		d.Set("encryption_entities", nil)
	}
	d.Set("etag", output.ETag)
	d.Set(names.AttrName, apiObject.Name)

	return diags
}

func resourceFieldLevelEncryptionProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	apiObject := &awstypes.FieldLevelEncryptionProfileConfig{
		CallerReference: aws.String(d.Get("caller_reference").(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_entities"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		apiObject.EncryptionEntities = expandEncryptionEntities(v.([]any)[0].(map[string]any))
	}

	input := &cloudfront.UpdateFieldLevelEncryptionProfileInput{
		FieldLevelEncryptionProfileConfig: apiObject,
		Id:                                aws.String(d.Id()),
		IfMatch:                           aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateFieldLevelEncryptionProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Field-level Encryption Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFieldLevelEncryptionProfileRead(ctx, d, meta)...)
}

func resourceFieldLevelEncryptionProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Field-level Encryption Profile: (%s)", d.Id())
	input := cloudfront.DeleteFieldLevelEncryptionProfileInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}
	_, err := conn.DeleteFieldLevelEncryptionProfile(ctx, &input)

	if errs.IsA[*awstypes.NoSuchFieldLevelEncryptionProfile](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Field-level Encryption Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findFieldLevelEncryptionProfileByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetFieldLevelEncryptionProfileOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionProfileInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionProfile(ctx, input)

	if errs.IsA[*awstypes.NoSuchFieldLevelEncryptionProfile](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FieldLevelEncryptionProfile == nil || output.FieldLevelEncryptionProfile.FieldLevelEncryptionProfileConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandEncryptionEntities(tfMap map[string]any) *awstypes.EncryptionEntities {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EncryptionEntities{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandEncryptionEntityItems(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandEncryptionEntity(tfMap map[string]any) *awstypes.EncryptionEntity {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EncryptionEntity{}

	if v, ok := tfMap["field_patterns"].([]any); ok && len(v) > 0 {
		apiObject.FieldPatterns = expandFieldPatterns(v[0].(map[string]any))
	}

	if v, ok := tfMap["provider_id"].(string); ok && v != "" {
		apiObject.ProviderId = aws.String(v)
	}

	if v, ok := tfMap["public_key_id"].(string); ok && v != "" {
		apiObject.PublicKeyId = aws.String(v)
	}

	return apiObject
}

func expandEncryptionEntityItems(tfList []any) []awstypes.EncryptionEntity {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EncryptionEntity

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandEncryptionEntity(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFieldPatterns(tfMap map[string]any) *awstypes.FieldPatterns {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FieldPatterns{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringValueSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func flattenEncryptionEntities(apiObject *awstypes.EncryptionEntities) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenEncryptionEntityItems(v)
	}

	return tfMap
}

func flattenEncryptionEntity(apiObject *awstypes.EncryptionEntity) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := flattenFieldPatterns(apiObject.FieldPatterns); len(v) > 0 {
		tfMap["field_patterns"] = []any{v}
	}

	if v := apiObject.ProviderId; v != nil {
		tfMap["provider_id"] = aws.ToString(v)
	}

	if v := apiObject.PublicKeyId; v != nil {
		tfMap["public_key_id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEncryptionEntityItems(apiObjects []awstypes.EncryptionEntity) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		if v := flattenEncryptionEntity(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenFieldPatterns(apiObject *awstypes.FieldPatterns) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = v
	}

	return tfMap
}

// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazoncloudfront.html#amazoncloudfront-resources-for-iam-policies.
func fieldLevelEncryptionProfileARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.GlobalARN(ctx, "cloudfront", "field-level-encryption-profile/"+id)
}
