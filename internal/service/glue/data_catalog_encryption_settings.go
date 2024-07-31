// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_data_catalog_encryption_settings")
func ResourceDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataCatalogEncryptionSettingsPut,
		ReadWithoutTimeout:   resourceDataCatalogEncryptionSettingsRead,
		UpdateWithoutTimeout: resourceDataCatalogEncryptionSettingsPut,
		DeleteWithoutTimeout: resourceDataCatalogEncryptionSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"data_catalog_encryption_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_password_encryption": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"return_connection_password_encrypted": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"encryption_at_rest": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog_encryption_mode": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CatalogEncryptionMode](),
									},
									"catalog_encryption_service_role": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"sse_aws_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
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

func resourceDataCatalogEncryptionSettingsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	input := &glue.PutDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(catalogID),
	}

	if v, ok := d.GetOk("data_catalog_encryption_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataCatalogEncryptionSettings = expandDataCatalogEncryptionSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Putting Glue Data Catalog Encryption Settings: %+v", input)
	_, err := conn.PutDataCatalogEncryptionSettings(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Glue Data Catalog Encryption Settings (%s): %s", catalogID, err)
	}

	d.SetId(catalogID)

	return append(diags, resourceDataCatalogEncryptionSettingsRead(ctx, d, meta)...)
}

func resourceDataCatalogEncryptionSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	output, err := conn.GetDataCatalogEncryptionSettings(ctx, &glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Data Catalog Encryption Settings (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCatalogID, d.Id())
	if output.DataCatalogEncryptionSettings != nil {
		if err := d.Set("data_catalog_encryption_settings", []interface{}{flattenDataCatalogEncryptionSettings(output.DataCatalogEncryptionSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting data_catalog_encryption_settings: %s", err)
		}
	} else {
		d.Set("data_catalog_encryption_settings", nil)
	}

	return diags
}

func resourceDataCatalogEncryptionSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.PutDataCatalogEncryptionSettingsInput{
		CatalogId:                     aws.String(d.Id()),
		DataCatalogEncryptionSettings: &awstypes.DataCatalogEncryptionSettings{},
	}

	log.Printf("[DEBUG] Deleting Glue Data Catalog Encryption Settings: %+v", input)
	_, err := conn.PutDataCatalogEncryptionSettings(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Glue Data Catalog Encryption Settings (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataCatalogEncryptionSettings(tfMap map[string]interface{}) *awstypes.DataCatalogEncryptionSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataCatalogEncryptionSettings{}

	if v, ok := tfMap["connection_password_encryption"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConnectionPasswordEncryption = expandConnectionPasswordEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["encryption_at_rest"].([]interface{}); ok && len(v) > 0 {
		apiObject.EncryptionAtRest = expandEncryptionAtRest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandConnectionPasswordEncryption(tfMap map[string]interface{}) *awstypes.ConnectionPasswordEncryption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ConnectionPasswordEncryption{}

	if v, ok := tfMap["aws_kms_key_id"].(string); ok && v != "" {
		apiObject.AwsKmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["return_connection_password_encrypted"].(bool); ok {
		apiObject.ReturnConnectionPasswordEncrypted = v
	}

	return apiObject
}

func expandEncryptionAtRest(tfMap map[string]interface{}) *awstypes.EncryptionAtRest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EncryptionAtRest{}

	if v, ok := tfMap["catalog_encryption_mode"].(string); ok && v != "" {
		apiObject.CatalogEncryptionMode = awstypes.CatalogEncryptionMode(v)
	}

	if v, ok := tfMap["catalog_encryption_service_role"].(string); ok && v != "" {
		apiObject.CatalogEncryptionServiceRole = aws.String(v)
	}

	if v, ok := tfMap["sse_aws_kms_key_id"].(string); ok && v != "" {
		apiObject.SseAwsKmsKeyId = aws.String(v)
	}

	return apiObject
}

func flattenDataCatalogEncryptionSettings(apiObject *awstypes.DataCatalogEncryptionSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConnectionPasswordEncryption; v != nil {
		tfMap["connection_password_encryption"] = []interface{}{flattenConnectionPasswordEncryption(v)}
	}

	if v := apiObject.EncryptionAtRest; v != nil {
		tfMap["encryption_at_rest"] = []interface{}{flattenEncryptionAtRest(v)}
	}

	return tfMap
}

func flattenConnectionPasswordEncryption(apiObject *awstypes.ConnectionPasswordEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsKmsKeyId; v != nil {
		tfMap["aws_kms_key_id"] = aws.ToString(v)
	}

	tfMap["return_connection_password_encrypted"] = apiObject.ReturnConnectionPasswordEncrypted

	return tfMap
}

func flattenEncryptionAtRest(apiObject *awstypes.EncryptionAtRest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["catalog_encryption_mode"] = string(apiObject.CatalogEncryptionMode)

	if v := apiObject.CatalogEncryptionServiceRole; v != nil {
		tfMap["catalog_encryption_service_role"] = aws.ToString(v)
	}

	if v := apiObject.SseAwsKmsKeyId; v != nil {
		tfMap["sse_aws_kms_key_id"] = aws.ToString(v)
	}

	return tfMap
}
