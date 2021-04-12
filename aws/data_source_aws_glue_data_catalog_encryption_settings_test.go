package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsGlueDataCatalogEncryptionSettings_basic(t *testing.T) {
	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	datasourceName := "data.aws_glue_data_catalog_encryption_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, glue.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsGlueDataCatalogEncryptionSettingsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsGlueDataCatalogEncryptionSettingsCheck(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_password_encrypted", resourceName, "connection_password_encrypted"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_password_kms_key_arn", resourceName, "connection_password_kms_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "encryption_mode", resourceName, "encryption_mode"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_password_encrypted", resourceName, "connection_password_encrypted"),
				),
			},
		},
	})
}

func testAccDataSourceAwsGlueDataCatalogEncryptionSettingsCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsGlueDataCatalogEncryptionSettingsConfig() string {
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
  id = aws_glue_data_catalog_encryption_settings.test.id
}
`
}
