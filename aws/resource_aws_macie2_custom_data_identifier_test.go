package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsMacie2CustomDataIdentifier_Name_Generated(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigNameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsMacie2CustomDataIdentifier_disappears(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigNameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsMacie2CustomDataIdentifier_NamePrefix(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigNamePrefix(namePrefix, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsMacie2CustomDataIdentifier_WithClassificationJob(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := "mdbatlas-test" //os.Getenv("AWS_S3_BUCKET_NAME")
	accountID := "520983883852"   //os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigComplete(regex, accountID, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsMacie2CustomDataIdentifier_WithTags(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	accountID := os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigCompleteWithTags(regex, accountID, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName string, macie2Session *macie2.GetCustomDataIdentifierOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetCustomDataIdentifierInput{Id: aws.String(rs.Primary.ID)}

		resp, err := conn.GetCustomDataIdentifier(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie CustomDataIdentifier %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccCheckAwsMacie2CustomDataIdentifierDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_custom_data_identifier" {
			continue
		}

		input := &macie2.GetCustomDataIdentifierInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetCustomDataIdentifier(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie CustomDataIdentifier %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testaccawsmacieCustomDataIdentifierconfigNameGenerated(regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  regex = "%s"

  depends_on = [aws_macie2_account.test]
}
`, regex)
}

func testaccawsmacieCustomDataIdentifierconfigNamePrefix(name, regex string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

	resource "aws_macie2_custom_data_identifier" "test" {
		name_prefix = %[1]q
		regex = "%s"

		depends_on = [aws_macie2_account.test]
	}
`, name, regex)
}

func testaccawsmacieCustomDataIdentifierconfigComplete(regex, s3AccountID, s3Bucket string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  regex                  = "%[1]s"
  description            = "this a description"
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]

  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_classification_job" "test" {
  custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
  job_type                   = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = "%[2]s"
      buckets    = ["%[3]s"]
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
}
`, regex, s3AccountID, s3Bucket)
}

func testaccawsmacieCustomDataIdentifierconfigCompleteWithTags(regex, s3AccountID, s3Bucket string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  regex                  = "%[1]s"
  description            = "this a description"
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]
  tags = {
    Key = "value"
  }

  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_classification_job" "test" {
  custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
  job_type                   = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = "%[2]s"
      buckets    = ["%[3]s"]
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
}
`, regex, s3AccountID, s3Bucket)
}
