package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogLaunchTemplateConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_template_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchTemplateConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchTemplateConstraintExists(resourceName, &dco),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "TEMPLATE"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					// rules can come back set in any order
					resource.TestMatchResourceAttr(resourceName, "rule.0.name", regexp.MustCompile("^rule0.$")),
					resource.TestMatchResourceAttr(resourceName, "rule.1.name", regexp.MustCompile("^rule0.$")),
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

func TestAccAWSServiceCatalogLaunchTemplateConstraint_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_template_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchTemplateConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(saltedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchTemplateConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogLaunchTemplateConstraint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceCatalogLaunchTemplateConstraintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_launch_template_constraint" {
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

func testAccCheckServiceCatalogLaunchTemplateConstraintExists(resourceName string, describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig(saltedName string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchTemplateConstraintConfigRequirements(saltedName),
		`
resource "aws_servicecatalog_launch_template_constraint" "test" {
  description = "description"
  rule {
    name = "rule01"
    rule_condition = jsonencode({
     "Fn::Equals" = [
       {"Ref" = "Environment"},
       "test"
      ]
    })
    assertion {
     assert = jsonencode({
       "Fn::Contains" = [
         ["m1.small"],
         {"Ref" = "InstanceType"}
       ]
     })
     assert_description = "For the test environment, the instance type must be m1.small"
    }
  }
  rule {
   name = "rule02"
   rule_condition = jsonencode({
     "Fn::Equals" = [
       {"Ref" = "Environment"}, 
       "prod"
     ]
   })
   assertion {
     assert = jsonencode({
       "Fn::Contains" = [
         ["m1.large"],
         {"Ref" = "InstanceType"} 
       ]
     })
     assert_description = "For the prod environment, the instance type must be m1.large"
   }
  }
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
}
`)
	return config
}

func testAccAWSServiceCatalogLaunchTemplateConstraintConfigRequirements(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_role(saltedName),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolios(saltedName),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_product(saltedName),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolioProductAssociations(),
	)
}

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_role(saltedName string) string {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolios(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName)
}

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_product(saltedName string) string {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolioProductAssociations() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}
`
}
