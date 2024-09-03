// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataCatalogEncryptionSettings_basic(t *testing.T) {
	t.Skipf("Skipping aws_glue_data_catalog_encryption_settings tests due to potential KMS key corruption")

	ctx := acctest.Context(t)

	var settings awstypes.DataCatalogEncryptionSettings

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	keyResourceName := "aws_kms_key.test"
	roleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogEncryptionSettingsConfig_nonEncrypted(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_service_role", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataCatalogEncryptionSettingsConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_service_role", ""),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", keyResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccDataCatalogEncryptionSettingsConfig_encrypted_with_catalog_encryption_service_role(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_service_role", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", keyResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccDataCatalogEncryptionSettingsConfig_nonEncrypted(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_service_role", ""),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", ""),
				),
			},
		},
	})
}

func testAccCheckDataCatalogEncryptionSettingsExists(ctx context.Context, resourceName string, v *awstypes.DataCatalogEncryptionSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Data Catalog Encryption Settings ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := conn.GetDataCatalogEncryptionSettings(ctx, &glue.GetDataCatalogEncryptionSettingsInput{
			CatalogId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*v = *output.DataCatalogEncryptionSettings

		return nil
	}
}

func testAccDataCatalogEncryptionSettingsConfig_encrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_glue_data_catalog_encryption_settings" "test" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      aws_kms_key_id                       = aws_kms_key.test.arn
      return_connection_password_encrypted = true
    }

    encryption_at_rest {
      catalog_encryption_mode = "SSE-KMS"
      sse_aws_kms_key_id      = aws_kms_key.test.arn
    }
  }
}
`, rName)
}

func testAccDataCatalogEncryptionSettingsConfig_encrypted_with_catalog_encryption_service_role(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Action": ["sts:AssumeRole"],
    }
  ]
}
POLICY
}

resource "aws_glue_data_catalog_encryption_settings" "test" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      aws_kms_key_id                       = aws_kms_key.test.arn
      return_connection_password_encrypted = true
    }

    encryption_at_rest {
      catalog_encryption_mode         = "SSE-KMS"
      catalog_encryption_service_role = aws_iam_role.test.arn
      sse_aws_kms_key_id              = aws_kms_key.test.arn
    }
  }
}
`, rName)
}

func testAccDataCatalogEncryptionSettingsConfig_nonEncrypted() string {
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
`
}
