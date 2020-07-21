package aws

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogConstraint_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogConstraintExists(resourceName, &dco),
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
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var describeConstraintOutput servicecatalog.DescribeConstraintOutput
	var providers []*schema.Provider
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogConstraintExists(resourceName, &describeConstraintOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogConstraint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogConstraint_updateDescription(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "description")),
			},
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, "-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "description-2")),
			},
		},
	})
}

func TestAccAWSServiceCatalogConstraint_updateParameters(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	var dco servicecatalog.DescribeConstraintOutput
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogConstraintExists(resourceName, &dco),
					testAccCheckServiceCatalogConstraintParametersLocalRoleName(&dco, "testpath/"+saltedName)),
			},
			{
				Config: testAccAWSServiceCatalogConstraintConfig(saltedName, "-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogConstraintExists(resourceName, &dco),
					testAccCheckServiceCatalogConstraintParametersLocalRoleName(&dco, "testpath/"+saltedName+"-2")),
			},
		},
	})
}

func testAccCheckServiceCatalogConstraintParametersLocalRoleName(dco *servicecatalog.DescribeConstraintOutput, expectedLocalRoleName string) resource.TestCheckFunc {
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

func testAccCheckServiceCatalogConstraintExists(resourceName string, dco *servicecatalog.DescribeConstraintOutput) resource.TestCheckFunc {
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

func testAccAWSServiceCatalogConstraintConfig(saltedName string, suffix string) string {
	return composeConfig(
		testAccAWSServiceCatalogConstraintConfigRequirements(saltedName),
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
`, saltedName, suffix))
}

func testAccAWSServiceCatalogConstraintConfigRequirements(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogConstraintConfig_role(saltedName),
		testAccAWSServiceCatalogConstraintConfig_portfolio(saltedName),
		testAccAWSServiceCatalogConstraintConfig_product(saltedName),
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

func testAccAWSServiceCatalogConstraintConfig_product(saltedName string) string {
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

func testAccAWSServiceCatalogConstraintConfig_portfolio(saltedName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName)
}

func testAccAWSServiceCatalogConstraintConfig_role(saltedName string) string {
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
resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
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
  description = "%[1]s-2"
  path = "/testpath/"
  force_detach_policies = false
  max_session_duration = 3600
}
`, saltedName)
}
