package aws

import (
	"encoding/json"
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

func TestAccAWSServiceCatalogConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraint(resourceName, &dco),
					resource.TestCheckResourceAttrSet(resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrSet(resourceName, "product_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "LAUNCH"),
					resource.TestCheckResourceAttrSet(resourceName, "parameters"),
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

func TestAccAWSServiceCatalogConstraint_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckServiceCatalogConstraintDisappears(&describeConstraintOutput),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogConstraint_updateDescription(t *testing.T) {
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "description", "description")),
			},
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, "_updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_constraint.test", "description", "description_updated")),
			},
		},
	})
}

func TestAccAWSServiceCatalogConstraint_updateParameters(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	salt := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	var dco servicecatalog.DescribeConstraintOutput
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraint(resourceName, &dco),
					testAccCheckConstraintParametersLocalRoleName(&dco, "testpath/"+salt)),
			},
			{
				Config: testAccAWSServiceCatalogConstraintConfig(salt, "_updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraint(resourceName, &dco),
					testAccCheckConstraintParametersLocalRoleName(&dco, "testpath/"+salt+"_updated")),
			},
		},
	})
}

func testAccCheckConstraintParametersLocalRoleName(dco *servicecatalog.DescribeConstraintOutput, expectedLocalRoleName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		parameters := dco.ConstraintParameters
		var bytes []byte = []byte(*parameters)
		//json decode parameters
		type LaunchParameters struct {
			LocalRoleName string
			RoleArn       string
		}
		var launchParameters LaunchParameters
		err := json.Unmarshal(bytes, &launchParameters)
		if err != nil {
			return err
		}
		// expect key LocalRoleName
		if launchParameters.LocalRoleName == "" {
			return fmt.Errorf("parameter LocalRoleName is not set")
		}
		// expect value expectedLocalRoleName
		if launchParameters.LocalRoleName != expectedLocalRoleName {
			return fmt.Errorf("parameter LocalRoleName is not expected value: '%s' (expected: '%s')",
				launchParameters.LocalRoleName, expectedLocalRoleName)
		}
		return nil // no errors
	}
}

func testAccCheckConstraint(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAWSServiceCatalogConstraintConfig(salt string, tag string) string {
	return composeConfig(
		testAccAWSServiceCatalogConstraintConfigRequirements(salt),
		fmt.Sprintf(`
resource "aws_servicecatalog_constraint" "test" {
  description = "description%[2]s"
  parameters = <<EOF
{
  "LocalRoleName": "testpath/%[1]s%[2]s"
}
EOF
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
  type = "LAUNCH"
}
`, salt, tag))
}

func testAccAWSServiceCatalogConstraintConfigRequirements(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogConstraintConfig_role(salt),
		testAccAWSServiceCatalogConstraintConfig_portfolio(salt),
		testAccAWSServiceCatalogConstraintConfig_product(salt),
		testAccAWSServiceCatalogConstraintConfig_portfolioProductAssociation(),
	)
}

func testAccAWSServiceCatalogConstraintConfig_portfolioProductAssociation() string {
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}`
}

func testAccAWSServiceCatalogConstraintConfig_product(salt string) string {
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "test" {
  bucket        = "terraform-test-%[1]s"
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
}`, salt)
}

func testAccAWSServiceCatalogConstraintConfig_portfolio(salt string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-2"
  provider_name = "test-3"
}
`, salt)
}

func testAccAWSServiceCatalogConstraintConfig_role(salt string) string {
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
resource "aws_iam_role" "test_alternate" {
  name = "%[1]s_updated"
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
  description = "%[1]s_updated"
  path = "/testpath/"
  force_detach_policies = false
  max_session_duration = 3600
}
`, salt)
}

func testAccCheckServiceCatalogConstraintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_constraint" {
			continue // not our monkey
		}
		input := servicecatalog.DescribeConstraintInput{Id: aws.String(rs.Primary.ID)}
		_, err := conn.DescribeConstraint(&input)
		if err == nil {
			return fmt.Errorf("constraint still exists")
		}
	}
	return nil
}

func testAccCheckServiceCatalogConstraintExists(resourceName string, describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccCheckServiceCatalogConstraintDisappears(describeConstraintOutput *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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
			return fmt.Errorf("could not delete constraint: #{err}")
		}
		if err := waitForServiceCatalogConstraintDeletion(conn, aws.StringValue(constraintId)); err != nil {
			return err
		}
		return nil
	}
}

func waitForServiceCatalogConstraintDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
	input := servicecatalog.DescribeConstraintInput{Id: aws.String(id)}
	stateConf := resource.StateChangeConf{
		Pending:      []string{servicecatalog.StatusAvailable},
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
