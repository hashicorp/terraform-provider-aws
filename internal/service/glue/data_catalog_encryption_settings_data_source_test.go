// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataCatalogEncryptionSettingsDataSource_basic(t *testing.T) {
	t.Skipf("Skipping aws_glue_data_catalog_encryption_settings tests due to potential KMS key corruption")

	ctx := acctest.Context(t)
	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	dataSourceName := "data.aws_glue_data_catalog_encryption_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogEncryptionSettingsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_catalog_encryption_settings", resourceName, "data_catalog_encryption_settings"),
				),
			},
		},
	})
}

func testAccDataCatalogEncryptionSettingsDataSourceConfig_basic() string {
	return `
resource "aws_glue_data_catalog_encryption_settings" "test" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      return_connection_password_encrypted = false
    }

    encryption_at_rest {
      catalog_encryption_mode = "DISABLED"
    }
  }
}

data "aws_glue_data_catalog_encryption_settings" "test" {
  catalog_id = aws_glue_data_catalog_encryption_settings.test.id
}
`
}
