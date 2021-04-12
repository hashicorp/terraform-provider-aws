package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsGlueDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsGlueDataCatalogEncryptionSettingsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"connection_password_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"connection_password_kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsGlueDataCatalogEncryptionSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).glueconn
	input := &glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(d.Id()),
	}
	out, err := conn.GetDataCatalogEncryptionSettings(input)
	if err != nil {
		return diag.Errorf("Error reading Glue Data Catalog Encryption Settings: %s", err)
	}
	d.SetId(d.Id())
	d.Set("connection_password_encrypted", out.DataCatalogEncryptionSettings.ConnectionPasswordEncryption.ReturnConnectionPasswordEncrypted)
	d.Set("connection_password_kms_key_arn", out.DataCatalogEncryptionSettings.ConnectionPasswordEncryption.AwsKmsKeyId)
	d.Set("encryption_mode", out.DataCatalogEncryptionSettings.EncryptionAtRest.CatalogEncryptionMode)
	d.Set("connection_password_encrypted", out.DataCatalogEncryptionSettings.EncryptionAtRest.SseAwsKmsKeyId)
	return nil
}
