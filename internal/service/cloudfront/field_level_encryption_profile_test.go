// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontFieldLevelEncryptionProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		CheckDestroy:             testAccCheckFieldLevelEncryptionProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(ctx, t, resourceName, &profile),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "field-level-encryption-profile/{id}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "some comment"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_entities.0.items.*", map[string]string{
						"provider_id":              rName,
						"field_patterns.#":         "1",
						"field_patterns.0.items.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "encryption_entities.0.items.*.field_patterns.0.items.*", "DateOfBirth"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFieldLevelEncryptionProfileConfig_extended(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(ctx, t, resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "some other comment"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_entities.0.items.*", map[string]string{
						"provider_id":              rName,
						"field_patterns.#":         "1",
						"field_patterns.0.items.#": "2",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "encryption_entities.0.items.*.field_patterns.0.items.*", "DateOfBirth"),
					resource.TestCheckTypeSetElemAttr(resourceName, "encryption_entities.0.items.*.field_patterns.0.items.*", "FirstName"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccCloudFrontFieldLevelEncryptionProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		CheckDestroy:             testAccCheckFieldLevelEncryptionProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(ctx, t, resourceName, &profile),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudfront.ResourceFieldLevelEncryptionProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFieldLevelEncryptionProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_field_level_encryption_profile" {
				continue
			}

			_, err := tfcloudfront.FindFieldLevelEncryptionProfileByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Field-level Encryption Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFieldLevelEncryptionProfileExists(ctx context.Context, t *testing.T, n string, v *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindFieldLevelEncryptionProfileByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFieldLevelEncryptionProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some comment"
  name    = %[1]q

  encryption_entities {
    items {
      public_key_id = aws_cloudfront_public_key.test.id
      provider_id   = %[1]q

      field_patterns {
        items = ["DateOfBirth"]
      }
    }
  }
}
`, rName)
}

func testAccFieldLevelEncryptionProfileConfig_extended(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some other comment"
  name    = %[1]q

  encryption_entities {
    items {
      public_key_id = aws_cloudfront_public_key.test.id
      provider_id   = %[1]q

      field_patterns {
        items = ["FirstName", "DateOfBirth"]
      }
    }
  }
}
`, rName)
}
