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

func TestAccCloudFrontFieldLevelEncryptionProfile_basic(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		CheckDestroy:      testAccCheckFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccFieldLevelEncryptionProfileExtendedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		CheckDestroy:      testAccCheckFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFieldLevelEncryptionProfileExists(resourceName, &profile),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceFieldLevelEncryptionProfile(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceFieldLevelEncryptionProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFieldLevelEncryptionProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_profile" {
			continue
		}

		_, err := tfcloudfront.FindFieldLevelEncryptionProfileByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Field-level Encryption Profile %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFieldLevelEncryptionProfileExists(r string, v *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Field-level Encryption Profile ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		output, err := tfcloudfront.FindFieldLevelEncryptionProfileByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFieldLevelEncryptionProfileConfig(rName string) string {
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

func testAccFieldLevelEncryptionProfileExtendedConfig(rName string) string {
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
