package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchEventPermission_Basic(t *testing.T) {
	principal1 := "111111111111"
	principal2 := "*"
	statementID := acctest.RandomWithPrefix(t.Name())
	resourceName := "aws_cloudwatch_event_permission.test1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic("", statementID),
				ExpectError: regexp.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(".", statementID),
				ExpectError: regexp.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic("12345678901", statementID),
				ExpectError: regexp.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic("abcdefghijkl", statementID),
				ExpectError: regexp.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal1, ""),
				ExpectError: regexp.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal1, acctest.RandString(65)),
				ExpectError: regexp.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal1, " "),
				ExpectError: regexp.MustCompile(`must be one or more alphanumeric, hyphen, or underscore characters`),
			},
			{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal1, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "events:PutEvents"),
					resource.TestCheckResourceAttr(resourceName, "principal", principal1),
					resource.TestCheckResourceAttr(resourceName, "statement_id", statementID),
				),
			},
			{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal2, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "principal", principal2),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventPermission_Action(t *testing.T) {
	principal := "111111111111"
	statementID := acctest.RandomWithPrefix(t.Name())
	resourceName := "aws_cloudwatch_event_permission.test1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigAction("", principal, statementID),
				ExpectError: regexp.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigAction(acctest.RandString(65), principal, statementID),
				ExpectError: regexp.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigAction("events:", principal, statementID),
				ExpectError: regexp.MustCompile(`must be: events: followed by one or more alphabetic characters`),
			},
			{
				Config:      testAccCheckAwsCloudWatchEventPermissionResourceConfigAction("events:1", principal, statementID),
				ExpectError: regexp.MustCompile(`must be: events: followed by one or more alphabetic characters`),
			},
			{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigAction("events:PutEvents", principal, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "events:PutEvents"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventPermission_Import(t *testing.T) {
	principal := "123456789012"
	statementID := acctest.RandomWithPrefix(t.Name())
	resourceName := "aws_cloudwatch_event_permission.test1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudWatchEventPermissionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName),
				),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCloudWatchEventPermission_Multiple(t *testing.T) {
	principal1 := "111111111111"
	principal2 := "222222222222"
	statementID1 := acctest.RandomWithPrefix(t.Name())
	statementID2 := acctest.RandomWithPrefix(t.Name())
	resourceName1 := "aws_cloudwatch_event_permission.test1"
	resourceName2 := "aws_cloudwatch_event_permission.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal1, statementID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName1),
					resource.TestCheckResourceAttr(resourceName1, "principal", principal1),
					resource.TestCheckResourceAttr(resourceName1, "statement_id", statementID1),
				),
			},
			{
				Config: testAccCheckAwsCloudWatchEventPermissionResourceConfigMultiple(principal1, statementID1, principal2, statementID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventPermissionExists(resourceName1),
					testAccCheckCloudWatchEventPermissionExists(resourceName2),
					resource.TestCheckResourceAttr(resourceName1, "principal", principal1),
					resource.TestCheckResourceAttr(resourceName1, "statement_id", statementID1),
					resource.TestCheckResourceAttr(resourceName2, "principal", principal2),
					resource.TestCheckResourceAttr(resourceName2, "statement_id", statementID2),
				),
			},
		},
	})
}

func testAccCheckCloudWatchEventPermissionExists(pr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		debo, err := conn.DescribeEventBus(&events.DescribeEventBusInput{})
		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %s", pr, err.Error())
		}

		if debo.Policy == nil {
			return fmt.Errorf("Not found: %s", pr)
		}

		var policyDoc CloudWatchEventPermissionPolicyDoc
		err = json.Unmarshal([]byte(*debo.Policy), &policyDoc)
		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %s", pr, err.Error())
		}

		_, err = findCloudWatchEventPermissionPolicyStatementByID(&policyDoc, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCloudWatchEventPermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_permission" {
			continue
		}

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			input := events.DescribeEventBusInput{}

			debo, err := conn.DescribeEventBus(&input)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			if debo.Policy == nil {
				return nil
			}

			var policyDoc CloudWatchEventPermissionPolicyDoc
			err = json.Unmarshal([]byte(*debo.Policy), &policyDoc)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Reading CloudWatch Events permission '%s' failed: %s", rs.Primary.ID, err.Error()))
			}

			_, err = findCloudWatchEventPermissionPolicyStatementByID(&policyDoc, rs.Primary.ID)
			if err == nil {
				return resource.RetryableError(fmt.Errorf("CloudWatch Events permission exists: %s", rs.Primary.ID))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAwsCloudWatchEventPermissionResourceConfigBasic(principal, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test1" {
  principal    = "%[1]s"
  statement_id = "%[2]s"
}
`, principal, statementID)
}

func testAccCheckAwsCloudWatchEventPermissionResourceConfigAction(action, principal, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test1" {
  action       = "%[1]s"
  principal    = "%[2]s"
  statement_id = "%[3]s"
}
`, action, principal, statementID)
}

func testAccCheckAwsCloudWatchEventPermissionResourceConfigMultiple(principal1, statementID1, principal2, statementID2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test1" {
  principal    = "%[1]s"
  statement_id = "%[2]s"
}

resource "aws_cloudwatch_event_permission" "test2" {
  principal    = "%[3]s"
  statement_id = "%[4]s"
}
`, principal1, statementID1, principal2, statementID2)
}
