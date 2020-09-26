package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSServiceCatalogLaunchNotificationConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_notification_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName, &describeConstraintOutput),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-1$`, saltedName))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-2$`, saltedName))),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogLaunchNotificationConstraint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchNotificationConstraint_updateNotificationArns(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_notification_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchNotificationConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName, &describeConstraintOutput),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-1$`, saltedName))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-2$`, saltedName))),
				),
			},
			{
				// now add a third sns topic
				Config: testAccAWSServiceCatalogLaunchNotificationConstraintConfig_updated(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchNotificationConstraintExists(resourceName, &describeConstraintOutput),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "NOTIFICATION"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "3"),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.0", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-1$`, saltedName))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.1", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-2$`, saltedName))),
					resource.TestMatchResourceAttr(resourceName, "notification_arns.2", regexp.MustCompile(fmt.Sprintf(`^arn:[^:]+:sns:[^:]+:[^:]+:%[1]s-topic-3$`, saltedName))),
				),
			},
		},
	})
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

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig(saltedName string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(saltedName),
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

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_updated(saltedName string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(saltedName),
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

func testAccAWSServiceCatalogLaunchNotificationConstraintConfigRequirements(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_role(saltedName),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolios(saltedName),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_product(saltedName),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolioProductAssociations(),
		testAccAWSServiceCatalogLaunchNotificationConstraintConfig_sns_topic(saltedName),
	)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_role(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
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
`, saltedName)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolios(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_product(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
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
  name                = %[1]q
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = %[1]q
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    }
  }
}`, saltedName)
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_portfolioProductAssociations() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}
`
}

func testAccAWSServiceCatalogLaunchNotificationConstraintConfig_sns_topic(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test1" {
  name = "%[1]s-topic-1"
}
resource "aws_sns_topic" "test2" {
  name = "%[1]s-topic-2"
}
resource "aws_sns_topic" "test3" {
  name = "%[1]s-topic-3"
}
`, saltedName)
}
