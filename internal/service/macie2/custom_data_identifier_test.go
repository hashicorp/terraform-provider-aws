package macie2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
)

func testAccCustomDataIdentifier_basic(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_nameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "regex", regex),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`custom-data-identifier/.+`)),
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

func testAccCustomDataIdentifier_Name_Generated(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_nameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
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

func testAccCustomDataIdentifier_disappears(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_nameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					acctest.CheckResourceDisappears(acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCustomDataIdentifier_NamePrefix(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_namePrefix(namePrefix, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`custom-data-identifier/.+`)),
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

func testAccCustomDataIdentifier_WithClassificationJob(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "this is a description"
	descriptionUpdated := "this is a updated description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_complete(bucketName, regex, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`custom-data-identifier/.+`)),
				),
			},
			{
				Config: testAccCustomDataIdentifierConfig_complete(bucketName, regex, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`custom-data-identifier/.+`)),
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

func testAccCustomDataIdentifier_WithTags(t *testing.T) {
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomDataIdentifierDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_completeTags(bucketName, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "value3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key3", "value3"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`custom-data-identifier/.+`)),
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

func testAccCheckCustomDataIdentifierExists(resourceName string, macie2Session *macie2.GetCustomDataIdentifierOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn
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

func testAccCheckCustomDataIdentifierDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_custom_data_identifier" {
			continue
		}

		input := &macie2.GetCustomDataIdentifierInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetCustomDataIdentifier(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
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

func testAccCustomDataIdentifierConfig_nameGenerated(regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  regex = %[1]q

  depends_on = [aws_macie2_account.test]
}
`, regex)
}

func testAccCustomDataIdentifierConfig_namePrefix(name, regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  name_prefix = %[1]q
  regex       = %[2]q

  depends_on = [aws_macie2_account.test]
}
`, name, regex)
}

func testAccCustomDataIdentifierConfig_complete(bucketName, regex, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie2_custom_data_identifier" "test" {
  regex                  = %[2]q
  description            = %[3]q
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
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
}
`, bucketName, regex, description)
}

func testAccCustomDataIdentifierConfig_completeTags(bucketName, regex string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie2_custom_data_identifier" "test" {
  regex                  = %[2]q
  description            = "this a description"
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]
  tags = {
    Key  = "value"
    Key2 = "value2"
    Key3 = "value3"
  }

  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_classification_job" "test" {
  custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
  job_type                   = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
}
`, bucketName, regex)
}
