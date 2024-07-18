// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontFieldLevelEncryptionConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		CheckDestroy:             testAccCheckFieldLevelEncryptionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "some comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.*", map[string]string{
						names.AttrContentType: "application/x-www-form-urlencoded",
						names.AttrFormat:      "URLEncoded",
					}),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.forward_when_content_type_is_unknown", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_profile_is_unknown", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFieldLevelEncryptionConfigConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.*", map[string]string{
						names.AttrContentType: "application/x-www-form-urlencoded",
						names.AttrFormat:      "URLEncoded",
					}),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_profile_is_unknown", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "query_arg_profile_config.0.query_arg_profiles.0.items.*", map[string]string{
						"query_arg": "Arg1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "query_arg_profile_config.0.query_arg_profiles.0.items.*", map[string]string{
						"query_arg": "Arg2",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccCloudFrontFieldLevelEncryptionConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		CheckDestroy:             testAccCheckFieldLevelEncryptionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceFieldLevelEncryptionConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFieldLevelEncryptionConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_field_level_encryption_config" {
				continue
			}

			_, err := tfcloudfront.FindFieldLevelEncryptionConfigByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Field-level Encryption Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFieldLevelEncryptionConfigExists(ctx context.Context, n string, v *cloudfront.GetFieldLevelEncryptionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindFieldLevelEncryptionConfigByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFieldLevelEncryptionConfig_base(rName string) string {
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

func testAccFieldLevelEncryptionConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFieldLevelEncryptionConfig_base(rName), `
resource "aws_cloudfront_field_level_encryption_config" "test" {
  comment = "some comment"

  content_type_profile_config {
    forward_when_content_type_is_unknown = true

    content_type_profiles {
      items {
        content_type = "application/x-www-form-urlencoded"
        format       = "URLEncoded"
        profile_id   = aws_cloudfront_field_level_encryption_profile.test.id
      }
    }
  }

  query_arg_profile_config {
    forward_when_query_arg_profile_is_unknown = true
  }
}
`)
}

func testAccFieldLevelEncryptionConfigConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccFieldLevelEncryptionConfig_base(rName), `
resource "aws_cloudfront_field_level_encryption_config" "test" {
  comment = "some other comment"

  content_type_profile_config {
    forward_when_content_type_is_unknown = true

    content_type_profiles {
      items {
        content_type = "application/x-www-form-urlencoded"
        format       = "URLEncoded"
        profile_id   = aws_cloudfront_field_level_encryption_profile.test.id
      }
    }
  }

  query_arg_profile_config {
    forward_when_query_arg_profile_is_unknown = false

    query_arg_profiles {
      items {
        profile_id = aws_cloudfront_field_level_encryption_profile.test.id
        query_arg  = "Arg1"
      }

      items {
        profile_id = aws_cloudfront_field_level_encryption_profile.test.id
        query_arg  = "Arg2"
      }
    }
  }
}
`)
}
