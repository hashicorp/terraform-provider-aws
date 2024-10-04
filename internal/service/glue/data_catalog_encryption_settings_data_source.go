// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_data_catalog_encryption_settings")
func DataSourceDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDataCatalogEncryptionSettingsRead,
		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"data_catalog_encryption_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_password_encryption": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"return_connection_password_encrypted": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"encryption_at_rest": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog_encryption_mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"catalog_encryption_service_role": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"sse_aws_kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceDataCatalogEncryptionSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID := d.Get(names.AttrCatalogID).(string)
	output, err := conn.GetDataCatalogEncryptionSettings(ctx, &glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(catalogID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Data Catalog Encryption Settings (%s): %s", catalogID, err)
	}

	d.SetId(catalogID)
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
