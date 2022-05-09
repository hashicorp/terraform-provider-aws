package glue_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataCatalogEncryptionSettingsDataSource_basic(t *testing.T) {
	t.Skipf("Skipping aws_glue_data_catalog_encryption_settings tests")

	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	dataSourceName := "data.aws_glue_data_catalog_encryption_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogEncryptionSettingsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_catalog_encryption_settings", resourceName, "data_catalog_encryption_settings"),
				),
			},
		},
	})
}

func testAccDataCatalogEncryptionSettingsDataSourceConfig() string {
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
