package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAwsMacie2CustomDataIdentifier_basic(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	name := acctest.RandomWithPrefix("testacc-custom-data-identifier")
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigBasic(name, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
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
	nameCustom := acctest.RandomWithPrefix("testacc-custom-data-identifier")
	nameJob := acctest.RandomWithPrefix("testacc-custom-data-identifier")
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := "mdbatlas-test" //os.Getenv("AWS_S3_BUCKET_NAME")
	clientToken := acctest.RandString(10)
	accountID := "520983883852" //os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigComplete(clientToken, nameCustom, regex, nameJob, accountID, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "client_token"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
			},
		},
	})
}

func TestAccAwsMacie2CustomDataIdentifier_WithTags(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	nameCustom := acctest.RandomWithPrefix("testacc-custom-data-identifier")
	nameJob := acctest.RandomWithPrefix("testacc-custom-data-identifier")
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := "mdbatlas-test" //os.Getenv("AWS_S3_BUCKET_NAME")
	clientToken := acctest.RandString(10)
	accountID := "520983883852" //os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2CustomDataIdentifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieCustomDataIdentifierconfigCompleteWithTags(clientToken, nameCustom, regex, nameJob, accountID, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2CustomDataIdentifierExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "client_token"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
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
			return fmt.Errorf("macie2 CustomDataIdentifier %q does not exist", rs.Primary.ID)
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

		if isAWSErr(err, macie2.ErrCodeAccessDeniedException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie2 CustomDataIdentifier %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testaccawsmacieCustomDataIdentifierconfigBasic(name, regex string) string {
	return fmt.Sprintf(`
	resource "aws_macie2_account" "test" {}

	resource "aws_macie2_custom_data_identifier" "test" {
		name = "%s"
		regex = "%s"

		depends_on = [aws_macie2_account.test]
	}
`, name, regex)
}

func testaccawsmacieCustomDataIdentifierconfigComplete(clientToken, nameCustom, regex, nameJob, s3AccountID, s3Bucket string) string {
	return fmt.Sprintf(`
	resource "aws_macie2_account" "test" {
		client_token = "%[1]s"
	}
	
	resource "aws_macie2_custom_data_identifier" "test" {
		name = "%[2]s"
		regex = "%[3]s"
		client_token = aws_macie2_account.test.client_token
		description = "this a description"
		maximum_match_distance = 10
		keywords = ["test"]
		ignore_words = ["not testing"]

		depends_on = [aws_macie2_account.test]
	}

	resource "aws_macie2_classification_job" "test" {
		custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
		client_token = "%[1]s"
		job_type = "SCHEDULED"
		name = "%[4]s"
		s3_job_definition {
			bucket_definitions{
				account_id = "%[5]s"
				buckets = ["%[6]s"]
			}
		}
		schedule_frequency {
			daily_schedule = true
		}
		sampling_percentage = 100
		description = "test"
		initial_run = true
	}
`, clientToken, nameCustom, regex, nameJob, s3AccountID, s3Bucket)
}

func testaccawsmacieCustomDataIdentifierconfigCompleteWithTags(clientToken, nameCustom, regex, nameJob, s3AccountID, s3Bucket string) string {
	return fmt.Sprintf(`
	resource "aws_macie2_account" "test" {
		client_token = "%s"
	}

	resource "aws_macie2_custom_data_identifier" "test" {
		name = "%[2]s"
		regex = "%[3]s"
		client_token = aws_macie2_account.test.client_token
		description = "this a description"
		maximum_match_distance = 10
		keywords = ["test"]
		ignore_words = ["not testing"]
		tags = {
    		Key = "value"
		}

		depends_on = [aws_macie2_account.test]
	}

	resource "aws_macie2_classification_job" "test" {
		custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
		client_token = "%[1]s"
		job_type = "SCHEDULED"
		name = "%[4]s"
		s3_job_definition {
			bucket_definitions{
				account_id = "%[5]s"
				buckets = ["%[6]s"]
			}
		}
		schedule_frequency {
			daily_schedule = true
		}
		sampling_percentage = 100
		description = "test"
		initial_run = true
	}
`, clientToken, nameCustom, regex, nameJob, s3AccountID, s3Bucket)
}
