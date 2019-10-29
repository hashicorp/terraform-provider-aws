package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAwsGuardDutyFilter_basic(t *testing.T) {
	resourceName := "aws_guardduty_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_to_fail_1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterDoesNotExist(resourceName),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_to_fail_2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterDoesNotExist(resourceName),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "detector_id", "123456271278c0df5e089123480d8765"),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_full,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "detector_id", "123456271278c0df5e089123480d8765"),
					// add more fields to test!
				),
			},
		},
	})
}

func testAccAwsGuardDutyFilter_import(t *testing.T) {
	resourceName := "aws_guardduty_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_basic,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsGuardDutyFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_filter" {
			continue
		}

		input := &guardduty.GetFilterInput{
			FilterName: aws.String(rs.Primary.Attributes["filter_name"]),
			DetectorId: aws.String(rs.Primary.Attributes["detector_id"]),
		}

		_, err := conn.GetFilter(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected GuardDuty Filter to be destroyed, %s found", rs.Primary.Attributes["filter_name"])
	}

	return nil
}

func testAccCheckAwsGuardDutyFilterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAwsGuardDutyFilterDoesNotExist(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return nil
		}

		return fmt.Errorf("Not found: %s", name)
	}
}

const testAccGuardDutyFilterConfig_to_fail_1 = `
resource "aws_guardduty_filter" "test" {}`

const testAccGuardDutyFilterConfig_to_fail_2 = `
resource "aws_guardduty_filter" "test" {
	detector_id = "123456271278c0df5e089123480d8765"
	}`

const testAccGuardDutyFilterConfig_basic = `
resource "aws_guardduty_filter" "test" {
  detector_id = "123456271278c0df5e089123480d8765"
	name = "test-filter"
}`

const testAccGuardDutyFilterConfig_full = `
resource "aws_guardduty_filter" "test" {
  detector_id = "123456271278c0df5e089123480d8765"
	name = "test-filter"
	action      = "ARCHIVE"
	rank        = 2

  finding_criteria {
    criterion {
      field     = "region"
      values    = ["eu-west-1"]
      condition = "equals"
    }

    criterion {
      field     = "service.additionalInfo.threatListName"
      values    = ["some-threat", "another-threat"]
      condition = "not_equals"
    }

    criterion {
      field     = "updatedAt"
      values    = ["1570744740000"]
      condition = "less_than"
    }

    criterion {
      field     = "updatedAt"
      values    = ["1570744240000"]
      condition = "greater_than"
    }
  }
}`
