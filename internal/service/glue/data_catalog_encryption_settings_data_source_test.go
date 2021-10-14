package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlueDataCatalogEncryptionSettingsDataSource_basic(t *testing.T) {
	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	datasourceName := "data.aws_glue_data_catalog_encryption_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogEncryptionSettingsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataCatalogEncryptionSettingsCheckDataSource(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "data_catalog_encryption_settings", resourceName, "data_catalog_encryption_settings"),
				),
			},
		},
	})
}

func testAccDataCatalogEncryptionSettingsCheckDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
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
