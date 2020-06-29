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

func TestAccAWSServiceCatalogLaunchTemplateConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_launch_template_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogLaunchTemplateConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateConstraint(resourceName, &dco),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "TEMPLATE"),
					//TODO fields to check
					//resource.TestCheckResourceAttrSet(roleArnResourceName, "role_arn"),
					//resource.TestCheckResourceAttr(roleArnResourceName, "local_role_name", ""),
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
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogLaunchTemplateConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogLaunchTemplateConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckServiceCatalogLaunchTemplateConstraintDisappears(&describeConstraintOutput),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//func TestAccAWSServiceCatalogLaunchTemplateConstraint_updateParameters(t *testing.T) {
//	resourceName := "aws_servicecatalog_launch_template_constraint.test"
//	roleArnResourceName := resourceName + "_a_role_arn"
//	localRoleNameResourceName := resourceName + "_b_local_role_name"
//	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckServiceCatalogLaunchTemplateConstraintDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(salt),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttrSet(roleArnResourceName, "role_arn"),
//					resource.TestCheckResourceAttr(roleArnResourceName, "local_role_name", ""),
//
//					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "local_role_name"),
//					resource.TestCheckResourceAttr(localRoleNameResourceName, "role_arn", ""),
//				),
//			},
//			{
//				// now swap the local_role_name and role_arn on each launch role constraint
//				Config: testAccAWSServiceCatalogLaunchTemplateConstraintConfig(salt),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttrSet(roleArnResourceName, "local_role_name"),
//					resource.TestCheckResourceAttr(roleArnResourceName, "role_arn", ""),
//
//					resource.TestCheckResourceAttrSet(localRoleNameResourceName, "role_arn"),
//					resource.TestCheckResourceAttr(localRoleNameResourceName, "local_role_name", ""),
//				),
//			},
//		},
//	})
//}

func testAccCheckLaunchTemplateConstraint(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig(salt string) string {
	config := composeConfig(
		testAccAWSServiceCatalogLaunchTemplateConstraintConfigRequirements(salt),
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfigRequirements(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_role(salt),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolios(salt),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_product(salt),
		testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolioProductAssociations(),
	)
}

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_role(salt string) string {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolios(salt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "tfm-test-%[1]s-A"
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
}

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_product(salt string) string {
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

func testAccAWSServiceCatalogLaunchTemplateConstraintConfig_portfolioProductAssociations() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}
`
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

func testAccCheckServiceCatalogLaunchTemplateConstraintDisappears(describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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
		if err := waitForServiceCatalogLaunchTemplateConstraintDeletion(conn,
			aws.StringValue(constraintId)); err != nil {
			return err
		}
		return nil
	}
}

func waitForServiceCatalogLaunchTemplateConstraintDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
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
