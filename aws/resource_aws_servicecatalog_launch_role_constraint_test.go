package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogLaunchRoleConstraint_basic(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceNameA := "aws_servicecatalog_launch_role_constraint.test_a_role_arn"
	resourceNameB := "aws_servicecatalog_launch_role_constraint.test_b_local_role_name"
	var describeConstraintOutputA servicecatalog.DescribeConstraintOutput
	var describeConstraintOutputB servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameA, &describeConstraintOutputA),
					resource.TestCheckResourceAttrSet(resourceNameA, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceNameA, "product_id"),
					resource.TestCheckResourceAttr(resourceNameA, "description", "description"),
					resource.TestCheckResourceAttr(resourceNameA, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(resourceNameA, "role_arn"),
					resource.TestCheckResourceAttr(resourceNameA, "local_role_name", ""),

					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameB, &describeConstraintOutputB),
					resource.TestCheckResourceAttrSet(resourceNameB, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceNameB, "product_id"),
					resource.TestCheckResourceAttr(resourceNameB, "description", "description"),
					resource.TestCheckResourceAttr(resourceNameB, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(resourceNameB, "local_role_name"),
					resource.TestCheckResourceAttr(resourceNameB, "role_arn", ""),
				),
			},
			{
				ResourceName:      resourceNameA,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameB,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchRoleConstraint_disappears(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceNameA := "aws_servicecatalog_launch_role_constraint.test_a_role_arn"
	resourceNameB := "aws_servicecatalog_launch_role_constraint.test_b_local_role_name"
	var describeConstraintOutputA servicecatalog.DescribeConstraintOutput
	var describeConstraintOutputB servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameB, &describeConstraintOutputA),
					testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceNameB, &describeConstraintOutputB),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogLaunchRoleConstraint(), resourceNameA),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogLaunchRoleConstraint(), resourceNameB),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogLaunchRoleConstraint_updateParameters(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceNameA := "aws_servicecatalog_launch_role_constraint.test_a_role_arn"
	resourceNameB := "aws_servicecatalog_launch_role_constraint.test_b_local_role_name"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchRoleConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceNameA, "role_arn"),
					resource.TestCheckResourceAttr(resourceNameA, "local_role_name", ""),

					resource.TestCheckResourceAttrSet(resourceNameB, "local_role_name"),
					resource.TestCheckResourceAttr(resourceNameB, "role_arn", ""),
				),
			},
			{
				// now swap the local_role_name and role_arn on each launch role constraint
				Config: testAccAWSServiceCatalogLaunchRoleConstraintConfig_Parameters_update(saltedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceNameA, "local_role_name"),
					resource.TestCheckResourceAttr(resourceNameA, "role_arn", ""),

					resource.TestCheckResourceAttrSet(resourceNameB, "role_arn"),
					resource.TestCheckResourceAttr(resourceNameB, "local_role_name", ""),
				),
			},
		},
	})
}

func testAccCheckServiceCatalogLaunchRoleConstraintExists(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAWSServiceCatalogLaunchRoleConstraintConfig(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(saltedName),
		fmt.Sprintf(`
resource "aws_servicecatalog_launch_role_constraint" "test_a_role_arn" {
  description = "description"
  role_arn = aws_iam_role.test.arn
  portfolio_id = aws_servicecatalog_portfolio.test_a.id
  product_id = aws_servicecatalog_product.test.id
}
resource "aws_servicecatalog_launch_role_constraint" "test_b_local_role_name" {
  description = "description"
  local_role_name = "testpath/%[1]s"
  portfolio_id = aws_servicecatalog_portfolio.test_b.id
  product_id = aws_servicecatalog_product.test.id
}
`,
			saltedName))
}

// as above, but with each constraint having swapped local_role_name and role_arn parameters
func testAccAWSServiceCatalogLaunchRoleConstraintConfig_Parameters_update(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(saltedName),
		fmt.Sprintf(`
resource "aws_servicecatalog_launch_role_constraint" "test_a_role_arn" {
  description = "description"
  local_role_name = "testpath/%[1]s"
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
			saltedName))
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfigRequirements(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_role(saltedName),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolios(saltedName),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_product(saltedName),
		testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolioProductAssociations(),
	)
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_role(saltedName string) string {
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

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_portfolios(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test_a" {
  name          = "%[1]s-A"
  description   = "test-2"
  provider_name = "test-3"
}
resource "aws_servicecatalog_portfolio" "test_b" {
  name          = "%[1]s-B"
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName)
}

func testAccAWSServiceCatalogLaunchRoleConstraintConfig_product(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
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
