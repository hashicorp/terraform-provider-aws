package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogLaunchRoleConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_role_constraint.test"
	roleArnResourceName := resourceName + "_a_role_arn"
	localRoleNameResourceName := resourceName + "_b_local_role_name"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var roleArnDco servicecatalog.DescribeConstraintOutput
	var localRoleNameDco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchRoleConstraint(roleArnResourceName, &roleArnDco),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "product_id"),
					resource.TestCheckResourceAttr(roleArnResourceName, "description", "description"),
					resource.TestCheckResourceAttr(roleArnResourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(roleArnResourceName, "role_arn"),
					resource.TestCheckResourceAttr(roleArnResourceName, "local_role_name", ""),

					testAccCheckLaunchRoleConstraint(localRoleNameResourceName, &localRoleNameDco),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "product_id"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "description", "description"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "local_role_name"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "role_arn", ""),
				),
			},
			{
				ResourceName:      roleArnResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      localRoleNameResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchRoleConstraint_disappears(t *testing.T) {
	resourceNameA := "aws_servicecatalog_launch_role_constraint.test_a_role_arn"
	resourceNameB := "aws_servicecatalog_launch_role_constraint.test_b_local_role_name"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var describeConstraintOutputA servicecatalog.DescribeConstraintOutput
	var describeConstraintOutputB servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameA, &describeConstraintOutputA),
					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameB, &describeConstraintOutputB),
					testAccCheckServiceCatalogLaunchRoleConstraintDisappears(&describeConstraintOutputA),
					testAccCheckServiceCatalogLaunchRoleConstraintDisappears(&describeConstraintOutputB),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchRoleConstraint_updateParameters(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_role_constraint.test"
	roleArnResourceName := resourceName + "_a_role_arn"
	localRoleNameResourceName := resourceName + "_b_local_role_name"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(roleArnResourceName, "role_arn"),
					resource.TestCheckResourceAttr(roleArnResourceName, "local_role_name", ""),

					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "local_role_name"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "role_arn", ""),
				),
			},
			{
				// now swap the local_role_name and role_arn on each launch role constraint
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfigAlternate(salt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(roleArnResourceName, "local_role_name"),
					resource.TestCheckResourceAttr(roleArnResourceName, "role_arn", ""),

					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "role_arn"),
					resource.TestCheckResourceAttr(localRoleNameResourceName, "local_role_name", ""),
				),
			},
		},
	})
}

func testAccCheckLaunchRoleConstraint(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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
		resp, err := conn.DescribeConstraint(&input)
		if err != nil {
			return err
		}
		*dco = *resp
		return nil
	}
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfig(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(salt),
		fmt.Sprintf(`
resource "aws_servicecatalog_launch_role_constraint" "test_a_role_arn" {
  description = "description"
  role_arn = aws_iam_role.test.arn
  portfolio_id = aws_servicecatalog_portfolio.test_a.id
  product_id = aws_servicecatalog_product.test.id
}
resource "aws_servicecatalog_launch_role_constraint" "test_b_local_role_name" {
  description = "description"
  local_role_name = "testpath/tfm-test-%[1]s"
  portfolio_id = aws_servicecatalog_portfolio.test_b.id
  product_id = aws_servicecatalog_product.test.id
}
`,
			salt))
}

// as above, but with each constraint having swapped local_role_name and role_arn parameters
func testAccAWSServiceCatalogLaunchRoleConstraintConfigAlternate(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(salt),
		fmt.Sprintf(`
resource "aws_servicecatalog_launch_role_constraint" "test_a_role_arn" {
  description = "description"
  local_role_name = "testpath/tfm-test-%[1]s"
  portfolio_id = aws_servicecatalog_portfolio.test_a.id
  product_id = aws_servicecatalog_product.test.id
}
resource "aws_servicecatalog_launch_role_constraint" "test_b_local_role_name" {
  description = "description"
  role_arn = aws_iam_role.test.arn
  portfolio_id = aws_servicecatalog_portfolio.test_b.id
  product_id = aws_servicecatalog_product.test.id
}
`,
			salt))
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_role(salt),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolios(salt),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_product(salt),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolioProductAssociations(),
	)
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_role(salt string) string {
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

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolios(salt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test_a" {
  name          = "tfm-test-%[1]s-A"
  description   = "test-2"
  provider_name = "test-3"
}
resource "aws_servicecatalog_portfolio" "test_b" {
  name          = "tfm-test-%[1]s-B"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_product(salt string) string {
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
  key     = "test_templates_for_terraform_sc_dev1.json"
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

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolioProductAssociations() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test_a" {
    portfolio_id = aws_servicecatalog_portfolio.test_a.id
    product_id = aws_servicecatalog_product.test.id
}
resource "aws_servicecatalog_portfolio_product_association" "test_b" {
    portfolio_id = aws_servicecatalog_portfolio.test_b.id
    product_id = aws_servicecatalog_product.test.id
}
`
}

func testAccCheckServiceCatalogLaunchRoleConstraintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_launch_role_constraint" {
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

func testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceName string, describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccCheckServiceCatalogLaunchRoleConstraintDisappears(describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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
			return fmt.Errorf("could not delete launch role constraint: #{err}")
		}
		if err := waitForServiceCatalogLaunchRoleConstraintDeletion(conn,
			aws.StringValue(constraintId)); err != nil {
			return err
		}
		return nil
	}
}

func waitForServiceCatalogLaunchRoleConstraintDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
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
