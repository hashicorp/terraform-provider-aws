package aws

import (
	"fmt"
	"testing"

<<<<<<< HEAD
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
=======
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
)

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_basic(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
<<<<<<< HEAD
	keyResourceName := "aws_cloudfront_public_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
=======
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFront(t) },
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.provider_id", rName),
<<<<<<< HEAD
					resource.TestCheckResourceAttrPair(resourceName, "encryption_entities.0.public_key_id", keyResourceName, "id"),
=======
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.field_patterns.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.provider_id", rName),
<<<<<<< HEAD
					resource.TestCheckResourceAttrPair(resourceName, "encryption_entities.0.public_key_id", keyResourceName, "id"),
=======
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.field_patterns.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_disappears(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
<<<<<<< HEAD
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
=======
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFront(t) },
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
<<<<<<< HEAD
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudfrontFieldLevelEncryptionProfile(), resourceName),
=======
					testAccCheckCloudfrontFieldLevelEncryptionProfileDisappears(&profile),
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_profile" {
			continue
		}

<<<<<<< HEAD
		_, err := finder.FieldLevelEncryptionProfileByID(conn, rs.Primary.ID)
=======
		params := &cloudfront.GetFieldLevelEncryptionProfileInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetFieldLevelEncryptionProfile(params)
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
		if err == nil {
			return fmt.Errorf("cloudfront Field Level Encryption Profile was not deleted")
		}
	}

	return nil
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileExists(r string, profile *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

<<<<<<< HEAD
		resp, err := finder.FieldLevelEncryptionProfileByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error retrieving Cloudfront Field Level Encryption Profile: %w", err)
=======
		params := &cloudfront.GetFieldLevelEncryptionProfileInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetFieldLevelEncryptionProfile(params)
		if err != nil {
			return fmt.Errorf("Error retrieving Cloudfront Field Level Encryption Profile: %s", err)
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
		}

		*profile = *resp

		return nil
	}
}

<<<<<<< HEAD
=======
func testAccCheckCloudfrontFieldLevelEncryptionProfileDisappears(profile *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		params := &cloudfront.DeleteFieldLevelEncryptionProfileInput{
			Id:      profile.FieldLevelEncryptionProfile.Id,
			IfMatch: profile.ETag,
		}

		_, err := conn.DeleteFieldLevelEncryptionProfile(params)

		return err
	}
}

>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
func testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
<<<<<<< HEAD
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
=======
  encoded_key = "${file("test-fixtures/cloudfront-public-key.pem")}"
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some comment"
  name    = %[1]q

  encryption_entities {
<<<<<<< HEAD
    public_key_id  = aws_cloudfront_public_key.test.id
=======
    public_key_id  = "${aws_cloudfront_public_key.test.id}"
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
    provider_id    = %[1]q
    field_patterns = ["DateOfBirth"]
  }
}
`, rName)
}

func testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
<<<<<<< HEAD
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
=======
  encoded_key = "${file("test-fixtures/cloudfront-public-key.pem")}"
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some other comment"
  name    = %[1]q

  encryption_entities {
<<<<<<< HEAD
    public_key_id  = aws_cloudfront_public_key.test.id
=======
    public_key_id  = "${aws_cloudfront_public_key.test.id}"
>>>>>>> 7e2a4846f (add support for cloudfront field level encryption profile)
    provider_id    = %[1]q
    field_patterns = ["FirstName", "DateOfBirth"]
  }
}
`, rName)
}
