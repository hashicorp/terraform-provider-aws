package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSServiceCatalogProvisionedProduct_basic(t *testing.T) {
	salt := acctest.RandString(5)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	var describeProvisionedProductOutput servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// need multiple independent providers for assume-role not to leak
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(salt),
			},
			{
				// provisioning with assume_role needs to be run in a second step
				// because a provider can only assume a role existing before its definition.
				// thus we do our config in two steps.
				// see https://github.com/hashicorp/terraform/issues/2430

				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput),
					testAccCheckAwsServiceCatalogProvisionedProductStandardFields(resourceName, &describeProvisionedProductOutput, salt),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalog", regexp.MustCompile(`stack/.+/pp-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", "tfm-sc-test-pp-"+salt),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			/* TODO import scaffolding doesn't work breaks because the custom provider it needs isn't available;
			   not sure why because the provider is defined in the template (test works without this block).
			   {
			       ResourceName:      resourceName,
			       ImportState:       true,
			       ImportStateVerify: true,
			   },
			*/
		},
	})
}

func TestAccAWSServiceCatalogProvisionedProduct_disappears(t *testing.T) {
	salt := acctest.RandString(5)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	var describeProvisionedProductOutput servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(salt),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(salt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput),
					testAccCheckAwsServiceCatalogProvisionedProductDisappears(&describeProvisionedProductOutput),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisionedProduct_tags(t *testing.T) {
	salt := acctest.RandString(5)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	var describeProvisionedProductOutput1, describeProvisionedProductOutput2, describeProvisionedProductOutput3 servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(salt),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(salt, "key1=\"value1\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(salt, "key1=\"value1updated\" \n key2=\"value2\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					func(s *terraform.State) error {
						if *describeProvisionedProductOutput1.ProvisionedProductDetail.Id == *describeProvisionedProductOutput2.ProvisionedProductDetail.Id {
							return fmt.Errorf("Provisioned product ID should have changed as tags ForceNew")
						}
						return nil
					},
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(salt, "key2=\"value2\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					func(s *terraform.State) error {
						if *describeProvisionedProductOutput2.ProvisionedProductDetail.Id == *describeProvisionedProductOutput3.ProvisionedProductDetail.Id {
							return fmt.Errorf("Provisioned product ID should have changed as tags ForceNew")
						}
						return nil
					},
				),
			},
		},
	})
}

// TODO parameters test

func testAccCheckAwsServiceCatalogProvisionedProductExists(pr string, describeProvisionedProductOutput *servicecatalog.DescribeProvisionedProductOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := servicecatalog.DescribeProvisionedProductInput{}
		input.Id = aws.String(rs.Primary.ID)

		output, err := conn.DescribeProvisionedProduct(&input)
		if err != nil {
			return err
		}
		*describeProvisionedProductOutput = *output

		return nil
	}
}

func testAccCheckAwsServiceCatalogProvisionedProductStandardFields(resourceName string, describeProvisionedProductOutput *servicecatalog.DescribeProvisionedProductOutput, salt string) resource.TestCheckFunc {
	expectedPPName := "tfm-sc-test-pp-" + salt
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", expectedPPName),
		resource.TestCheckResourceAttr(resourceName, "status", servicecatalog.StatusAvailable),
		func(s *terraform.State) error {
			if *describeProvisionedProductOutput.ProvisionedProductDetail.Name != expectedPPName {
				return fmt.Errorf("resource '%s' does not have expected name: '%s' vs '%s'", resourceName, *describeProvisionedProductOutput.ProvisionedProductDetail.Name, expectedPPName)
			}
			return nil
		},
	)
}

func testAccCheckAwsServiceCatalogProvisionedProductDisappears(describeProvisionedProductOutput *servicecatalog.DescribeProvisionedProductOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		input := servicecatalog.TerminateProvisionedProductInput{
			ProvisionedProductId: describeProvisionedProductOutput.ProvisionedProductDetail.Id,
		}
		// not available on servicecatalog, but returned here if under change
		errCodeValidationException := "ValidationException"
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.TerminateProvisionedProduct(&input)
			if err != nil {
				if isAWSErr(err, servicecatalog.ErrCodeResourceInUseException, "") || isAWSErr(err, errCodeValidationException, "") {
					// delay and retry, other things eg associations might still be getting deleted
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not terminate provisioned product: %s", err)
		}
		if err := waitForServiceCatalogProvisionedProductDeletion(conn, aws.StringValue(describeProvisionedProductOutput.ProvisionedProductDetail.Id)); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckAwsServiceCatalogProvisionedProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_provisioned_product" {
			continue
		}
		input := servicecatalog.DescribeProvisionedProductInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribeProvisionedProduct(&input)
		if err != nil {
			if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("provisioned product still exists")
	}

	return nil
}

func testAccAWSServiceCatalogProvisionedProductConfig_portfolio(salt string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, "tfm-test-"+salt)
}

func testAccAWSServiceCatalogProvisionedProductConfig_role(salt string) string {
	roleName := "tfm-sc-tester-" + salt
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": { "AWS": "*" },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, roleName)
}

func testAccAWSServiceCatalogProvisionedProductConfig_portfolio_principal_association() string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_principal_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    principal_arn = aws_iam_role.test.arn
}
`)
}

func testAccAWSServiceCatalogProvisionedProductConfig_policy() string {
	return fmt.Sprintf(`

resource "aws_iam_role_policy" "test" {
  name = "test_policy"
  role = aws_iam_role.test.id

  policy = <<-EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Action": [
          "servicecatalog:*",
          "cloudformation:*",
          "s3:*"
        ],
        "Effect": "Allow",
        "Resource": "*"
      }
    ]
  }
  EOF
}
`)
}

func testAccAWSServiceCatalogProvisionedProductConfig_portfolio_product_association() string {
	// based testAccAWSServiceCatalogPortfolioProductAssociationConfig_basic
	return `
resource "aws_servicecatalog_portfolio_product_association" "test" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}
`
}

func testAccAWSServiceCatalogProvisionedProductConfig_product(salt string) string {
	// based on testAccAWSServiceCatalogProductConfig_basic
	resourceName := "aws_servicecatalog_product.test"

	thisResourceParts := strings.Split(resourceName, ".")
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "bucket-%[3]s"
  region        = data.aws_region.current.name
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
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

resource "%[1]s" "%[2]s" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "product-%[3]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "pa-%[3]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

}
`, thisResourceParts[0], thisResourceParts[1], salt)
}

func testAccAWSServiceCatalogProvisionedProductConfig_step1(salt string) string {
	return composeConfig(
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio(salt),
		testAccAWSServiceCatalogProvisionedProductConfig_role(salt),
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio_principal_association(),
		testAccAWSServiceCatalogProvisionedProductConfig_policy(),
		testAccAWSServiceCatalogProvisionedProductConfig_product(salt),
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio_product_association(),
	)
}

func testAccAWSServiceCatalogProvisionedProductConfig_step2(salt string, tags string) string {
	provisionedProductName := "tfm-sc-test-pp-" + salt
	return composeConfig(
		testAccAWSServiceCatalogProvisionedProductConfig_step1(salt),
		fmt.Sprintf(`

provider "aws" {
  alias               = "product-allowed-role"
  assume_role {
    role_arn          = aws_iam_role.test.arn
    session_name      = "tfm-sc-testing"
    external_id       = "tfm-sc-testing"
  }
}
`)+fmt.Sprintf(`

resource "aws_servicecatalog_provisioned_product" "test" {
    provider = aws.product-allowed-role
    provisioned_product_name = "%s"
    product_id               = aws_servicecatalog_product.test.id
    provisioning_artifact_id = aws_servicecatalog_product.test.provisioning_artifact[0].id
    depends_on = [
      aws_iam_role_policy.test,
      aws_servicecatalog_portfolio_product_association.test,
      aws_servicecatalog_portfolio_principal_association.test,
    ]
    tags = {
      %s
    }
}
`, provisionedProductName, tags))
}
