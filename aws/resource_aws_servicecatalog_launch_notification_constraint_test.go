package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogLaunchNotificationConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_notification_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchNotificationConstraint(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-1-%[1]s$`, salt))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-2-%[1]s$`, salt))),
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

func TestAccAWSServiceCatalogLaunchNotificationConstraint_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_notification_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckServiceCatalogLaunchNotificationConstraintDisappears(&describeConstraintOutput),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchNotificationConstraint_updateNotificationArns(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_notification_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchNotificationConstraint(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-1-%[1]s$`, salt))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-2-%[1]s$`, salt))),
				),
			},
			{
				// now add a third sns topic
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig_updated(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchNotificationConstraint(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "3"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-1-%[1]s$`, salt))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-2-%[1]s$`, salt))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.2", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:topic-3-%[1]s$`, salt))),
				),
			},
		},
	})
}

func testAccCheckLaunchNotificationConstraint(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("constraint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		input := servicecatalog.DescribeConstraintInput{
			Id: aws.String(rs.Primary.ID),
		}
		conn := testAccProvider.Meta().(*AWSClient).scconn
		_, err := conn.DescribeConstraint(&input)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig(salt string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(salt),
		`
resource "aws_servicecatalog_launch_notification_constraint" "test" {
  description = "description"
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
  notification_arns = [
    aws_sns_topic.test1.arn,
    aws_sns_topic.test2.arn
  ]
}
`)
	return config
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_updated(salt string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(salt),
		`
resource "aws_servicecatalog_launch_notification_constraint" "test" {
  description = "description"
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
  notification_arns = [
    aws_sns_topic.test1.arn,
    aws_sns_topic.test2.arn,
    aws_sns_topic.test3.arn
  ]
}
`)
	return config
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_role(salt),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolios(salt),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_product(salt),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolioProductAssociations(),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_sns_topic(salt),
	)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_role(salt string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tfm-test-%[1]s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "servicecatalog.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
  description = %[1]q
  path = "/testpath/"
  force_detach_policies = false
  max_session_duration = 3600
}
`, salt)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolios(salt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "tfm-test-%[1]s-A"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_product(salt string) string {
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "test" {
  bucket        = "tfm-test-%[1]s"
  region        = data.aws_region.current.name
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "test_notifications_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_product" "test" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "tfm-test-%[1]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "tfm-test-%[1]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    }
  }
}`, salt)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolioProductAssociations() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}
`
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_sns_topic(salt string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test1" {
  name = "topic-1-%[1]s"
}
resource "aws_sns_topic" "test2" {
  name = "topic-2-%[1]s"
}
resource "aws_sns_topic" "test3" {
  name = "topic-3-%[1]s"
}
`, salt)
}

func testAccCheckServiceCatalogLaunchNotificationConstraintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_launch_notification_constraint" {
			continue // not our monkey
		}
		input := servicecatalog.DescribeConstraintInput{Id: aws.String(rs.Primary.ID)}
		_, err := conn.DescribeConstraint(&input)
		if err == nil {
			return fmt.Errorf("constraint still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName string, describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		input := servicecatalog.DescribeConstraintInput{Id: aws.String(rs.Primary.ID)}
		conn := testAccProvider.Meta().(*AWSClient).scconn
		constraint, err := conn.DescribeConstraint(&input)
		if err != nil {
			return err
		}
		*describeConstraintOutput = *constraint
		return nil
	}
}

func testAccCheckServiceCatalogLaunchNotificationConstraintDisappears(describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		constraintId := describeConstraintOutput.ConstraintDetail.ConstraintId
		input := servicecatalog.DeleteConstraintInput{Id: constraintId}
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteConstraint(&input)
			if err != nil {
				if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") ||
					isAWSErr(err, servicecatalog.ErrCodeInvalidParametersException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not delete launch notification constraint: #{err}")
		}
		if err := waitForServiceCatalogLaunchNotificationConstraintDeletion(conn,
			aws.StringValue(constraintId)); err != nil {
			return err
		}
		return nil
	}
}

func waitForServiceCatalogLaunchNotificationConstraintDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
	input := servicecatalog.DescribeConstraintInput{Id: aws.String(id)}
	stateConf := resource.StateChangeConf{
		Pending:      []string{"AVAILABLE"},
		Target:       []string{""},
		Timeout:      5 * time.Minute,
		PollInterval: 20 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeConstraint(&input)
			if err != nil {
				if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException,
					fmt.Sprintf("Constraint %s not found.", id)) {
					return 42, "", nil
				}
				return 42, "", err
			}
			return resp, aws.StringValue(resp.Status), nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}
