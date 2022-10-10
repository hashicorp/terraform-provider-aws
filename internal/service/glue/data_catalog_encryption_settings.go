package glue

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataCatalogEncryptionSettingsPut,
		Read:   resourceDataCatalogEncryptionSettingsRead,
		Update: resourceDataCatalogEncryptionSettingsPut,
		Delete: resourceDataCatalogEncryptionSettingsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(glue.CatalogEncryptionMode_Values(), false),
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

func resourceDataCatalogEncryptionSettingsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	input := &glue.PutDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(catalogID),
	}

	if v, ok := d.GetOk("data_catalog_encryption_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataCatalogEncryptionSettings = expandDataCatalogEncryptionSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Putting Glue Data Catalog Encryption Settings: %s", input)
	_, err := conn.PutDataCatalogEncryptionSettings(input)

	if err != nil {
		return fmt.Errorf("error putting Glue Data Catalog Encryption Settings (%s): %w", catalogID, err)
	}

	d.SetId(catalogID)

	return resourceDataCatalogEncryptionSettingsRead(d, meta)
}

func resourceDataCatalogEncryptionSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	output, err := conn.GetDataCatalogEncryptionSettings(&glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading Glue Data Catalog Encryption Settings (%s): %w", d.Id(), err)
	}

	d.Set("catalog_id", d.Id())
	if output.DataCatalogEncryptionSettings != nil {
		if err := d.Set("data_catalog_encryption_settings", []interface{}{flattenDataCatalogEncryptionSettings(output.DataCatalogEncryptionSettings)}); err != nil {
			return fmt.Errorf("error setting data_catalog_encryption_settings: %w", err)
		}
	} else {
		d.Set("data_catalog_encryption_settings", nil)
	}

	return nil
}

func resourceDataCatalogEncryptionSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	input := &glue.PutDataCatalogEncryptionSettingsInput{
		CatalogId:                     aws.String(d.Id()),
		DataCatalogEncryptionSettings: &glue.DataCatalogEncryptionSettings{},
	}

	log.Printf("[DEBUG] Deleting Glue Data Catalog Encryption Settings: %s", input)
	_, err := conn.PutDataCatalogEncryptionSettings(input)

	if err != nil {
		return fmt.Errorf("error putting Glue Data Catalog Encryption Settings (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDataCatalogEncryptionSettings(tfMap map[string]interface{}) *glue.DataCatalogEncryptionSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.DataCatalogEncryptionSettings{}

	if v, ok := tfMap["connection_password_encryption"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConnectionPasswordEncryption = expandConnectionPasswordEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["encryption_at_rest"].([]interface{}); ok && len(v) > 0 {
		apiObject.EncryptionAtRest = expandEncryptionAtRest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandConnectionPasswordEncryption(tfMap map[string]interface{}) *glue.ConnectionPasswordEncryption {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.ConnectionPasswordEncryption{}

	if v, ok := tfMap["aws_kms_key_id"].(string); ok && v != "" {
		apiObject.AwsKmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["return_connection_password_encrypted"].(bool); ok {
		apiObject.ReturnConnectionPasswordEncrypted = aws.Bool(v)
	}

	return apiObject
}

func expandEncryptionAtRest(tfMap map[string]interface{}) *glue.EncryptionAtRest {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.EncryptionAtRest{}

	if v, ok := tfMap["catalog_encryption_mode"].(string); ok && v != "" {
		apiObject.CatalogEncryptionMode = aws.String(v)
	}

	if v, ok := tfMap["sse_aws_kms_key_id"].(string); ok && v != "" {
		apiObject.SseAwsKmsKeyId = aws.String(v)
	}

	return apiObject
}

func flattenDataCatalogEncryptionSettings(apiObject *glue.DataCatalogEncryptionSettings) map[string]interface{} {
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

func flattenConnectionPasswordEncryption(apiObject *glue.ConnectionPasswordEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsKmsKeyId; v != nil {
		tfMap["aws_kms_key_id"] = aws.StringValue(v)
	}

	if v := apiObject.ReturnConnectionPasswordEncrypted; v != nil {
		tfMap["return_connection_password_encrypted"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenEncryptionAtRest(apiObject *glue.EncryptionAtRest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogEncryptionMode; v != nil {
		tfMap["catalog_encryption_mode"] = aws.StringValue(v)
	}

	if v := apiObject.SseAwsKmsKeyId; v != nil {
		tfMap["sse_aws_kms_key_id"] = aws.StringValue(v)
	}

	return tfMap
}
