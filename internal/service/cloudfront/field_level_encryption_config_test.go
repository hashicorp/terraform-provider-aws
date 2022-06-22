package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFrontFieldLevelEncryptionConfig_basic(t *testing.T) {
	var v cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		CheckDestroy:      testAccCheckFieldLevelEncryptionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.*", map[string]string{
						"content_type": "application/x-www-form-urlencoded",
						"format":       "URLEncoded",
					}),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.forward_when_content_type_is_unknown", "true"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_profile_is_unknown", "true"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.#", "0"),
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
					testAccCheckFieldLevelEncryptionConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "comment", "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "content_type_profile_config.0.content_type_profiles.0.items.*", map[string]string{
						"content_type": "application/x-www-form-urlencoded",
						"format":       "URLEncoded",
					}),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_profile_is_unknown", "false"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profiles.0.items.#", "2"),
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
	var v cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		CheckDestroy:      testAccCheckFieldLevelEncryptionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceFieldLevelEncryptionConfig(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceFieldLevelEncryptionConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFieldLevelEncryptionConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_config" {
			continue
		}

		_, err := tfcloudfront.FindFieldLevelEncryptionConfigByID(conn, rs.Primary.ID)

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

func testAccCheckFieldLevelEncryptionConfigExists(r string, v *cloudfront.GetFieldLevelEncryptionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Field-level Encryption Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		output, err := tfcloudfront.FindFieldLevelEncryptionConfigByID(conn, rs.Primary.ID)

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
