package guardduty_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccDetector_basic(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyDetectorConfig_basic1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/.+$")),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "FIFTEEN_MINUTES"),
				),
			},
		},
	})
}

func testAccDetector_tags(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDetectorConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDetector_datasources_s3logs(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
				),
			},
		},
	})
}

func testAccCheckDetectorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_detector" {
			continue
		}

		input := &guardduty.GetDetectorInput{
			DetectorId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDetector(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected GuardDuty Detector to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDetectorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccGuardDutyDetectorConfig_basic1 = `
resource "aws_guardduty_detector" "test" {}
`

const testAccGuardDutyDetectorConfig_basic2 = `
resource "aws_guardduty_detector" "test" {
  enable = false
}
`

const testAccGuardDutyDetectorConfig_basic3 = `
resource "aws_guardduty_detector" "test" {
  enable = true
}
`

const testAccGuardDutyDetectorConfig_basic4 = `
resource "aws_guardduty_detector" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
}
`

func testAccDetectorConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccDetectorConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDetectorConfig_datasourcesS3Logs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    s3_logs {
      enable = %[1]t
    }
  }
}
`, enable)
}
