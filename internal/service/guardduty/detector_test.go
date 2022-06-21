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
				Config: testAccDetectorConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/.+$")),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
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
				Config: testAccDetectorConfig_disable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
			},
			{
				Config: testAccDetectorConfig_enable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
			},
			{
				Config: testAccDetectorConfig_findingPublishingFrequency,
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDetectorConfig_tags1("key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_datasources_kubernetes_audit_logs(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_datasources_all(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesAll(true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
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
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) has empty ID", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn

		output, err := conn.GetDetector(&guardduty.GetDetectorInput{
			DetectorId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("GuardDuty Detector not found: %s", name)
		}

		return nil
	}
}

const testAccDetectorConfig_basic = `
resource "aws_guardduty_detector" "test" {}
`

const testAccDetectorConfig_disable = `
resource "aws_guardduty_detector" "test" {
  enable = false
}
`

const testAccDetectorConfig_enable = `
resource "aws_guardduty_detector" "test" {
  enable = true
}
`

const testAccDetectorConfig_findingPublishingFrequency = `
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

func testAccDetectorConfig_datasourcesKubernetesAuditLogs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesAll(enableK8s, enableS3 bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
    s3_logs {
      enable = %[2]t
    }
  }
}
`, enableK8s, enableS3)
}
