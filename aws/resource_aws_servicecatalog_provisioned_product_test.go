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

// NB: these tests are slow: each create or update to a provision product can take a minute or two;
// run with `go test -timeout 20m ./aws -run ServiceCatalogProvisionedProduct`

func TestAccAWSServiceCatalogProvisionedProduct_basic(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
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
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
			},
			{
				// provisioning a product needs to be run in a second step,
				// using a provider the assumes the role created in the first step.
				// this is because provisioning requires an explicit principal (group/role/user),
				// and we don't know what principals the test configuration has.
				// the solution used here is to create a role in step 1, then assume_role here.
				// but a provider can only assume a role existing before its definition - https://github.com/hashicorp/terraform/issues/2430 -
				// hence the need to do it in two steps.

				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput),
					testAccCheckAwsServiceCatalogProvisionedProductStandardFields(resourceName, &describeProvisionedProductOutput, saltedName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalog", regexp.MustCompile(`stack/.+/pp-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", saltedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			/*
				// the import test scaffolding doesn't seem to be compatible with the provider setup
				// (variable input, pointing at the role created in step 1) used in this test, as described above.
				// it gives an error:
				//   Error: config is invalid: Provider configuration not present: To work with aws_servicecatalog_provisioned_product.test its original provider configuration at provider.aws.product-allowed-role is required, but it has been removed. This occurs when a provider configuration is removed while objects created by that provider still exist in the state. Re-add the provider configuration to destroy aws_servicecatalog_provisioned_product.test, after which you can remove the provider configuration again.
				// import itself works fine, outwith testing, but here it's not clear how to configure the import test.
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
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_provisioned_product.test"
	var describeProvisionedProductOutput servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProvisionedProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisionedProduct_tags(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_provisioned_product.test"
	var describeProvisionedProductOutput1, describeProvisionedProductOutput2, describeProvisionedProductOutput3 servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName, "key1=\"value1\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName, "key1=\"value1updated\" \n key2=\"value2\""),
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
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName, "key2=\"value2\""),
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

func TestAccAWSServiceCatalogProvisionedProduct_ProvisioningParameters(t *testing.T) {
	saltedName := "tf-acc-test-" + acctest.RandString(5) // RandomWithPrefix exceeds max length 20
	resourceName := "aws_servicecatalog_provisioned_product.test_params"
	var describeProvisionedProductOutput1, describeProvisionedProductOutput2 servicecatalog.DescribeProvisionedProductOutput
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2_params(saltedName, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput1),
					resource.TestCheckResourceAttr(resourceName, "outputs.NumberWithRange", "42"),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_step2_params(saltedName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName, &describeProvisionedProductOutput2),
					resource.TestCheckResourceAttr(resourceName, "outputs.NumberWithRange", "60"),
					func(s *terraform.State) error {
						if *describeProvisionedProductOutput1.ProvisionedProductDetail.Id != *describeProvisionedProductOutput2.ProvisionedProductDetail.Id {
							return fmt.Errorf("Provisioned product ID should not have changed on parameters change")
						}
						return nil
					},
				),
			},
		},
	})
}

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

func testAccCheckAwsServiceCatalogProvisionedProductStandardFields(resourceName string, describeProvisionedProductOutput *servicecatalog.DescribeProvisionedProductOutput, saltedName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", saltedName),
		resource.TestCheckResourceAttr(resourceName, "status", servicecatalog.StatusAvailable),
		func(s *terraform.State) error {
			if aws.StringValue(describeProvisionedProductOutput.ProvisionedProductDetail.Name) != saltedName {
				return fmt.Errorf("resource '%s' does not have expected name: '%s' vs '%s'", resourceName, *describeProvisionedProductOutput.ProvisionedProductDetail.Name, saltedName)
			}
			return nil
		},
	)
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

func testAccAWSServiceCatalogProvisionedProductConfig_portfolio(saltedName string) string {
	// based on testAccAWSServiceCatalogPortfolioConfig_basic
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "test-2"
  provider_name = "test-3"
}
`, saltedName)
}

func testAccAWSServiceCatalogProvisionedProductConfig_role(saltedName string) string {
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
`, saltedName)
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

func testAccAWSServiceCatalogProvisionedProductConfig_product(saltedName string) string {
	// based on testAccAWSServiceCatalogProductConfig_basic
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
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
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

}
`, saltedName)
}

func testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName string) string {
	return composeConfig(
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio(saltedName),
		testAccAWSServiceCatalogProvisionedProductConfig_role(saltedName),
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio_principal_association(),
		testAccAWSServiceCatalogProvisionedProductConfig_policy(),
		testAccAWSServiceCatalogProvisionedProductConfig_product(saltedName),
		testAccAWSServiceCatalogProvisionedProductConfig_portfolio_product_association(),
	)
}

func testAccAWSServiceCatalogProvisionedProductConfig_provider() string {
	return `

provider "aws" {
  alias               = "product-allowed-role"
  assume_role {
    role_arn          = aws_iam_role.test.arn
    session_name      = "tfm-sc-testing"
    external_id       = "tfm-sc-testing"
  }
}
`
}

func testAccAWSServiceCatalogProvisionedProductConfig_step2(saltedName string, tags string) string {
	return composeConfig(
		testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
		testAccAWSServiceCatalogProvisionedProductConfig_provider(),
		fmt.Sprintf(`

resource "aws_servicecatalog_provisioned_product" "test" {
    provider                 = aws.product-allowed-role
    provisioned_product_name = %[1]q
    product_id               = aws_servicecatalog_product.test.id
    provisioning_artifact_id = aws_servicecatalog_product.test.provisioning_artifact[0].id
    depends_on = [
      aws_iam_role_policy.test,
      aws_servicecatalog_portfolio_product_association.test,
      aws_servicecatalog_portfolio_principal_association.test,
    ]
    tags = {
      %[2]s
    }
}
`, saltedName, tags))
}

func testAccAWSServiceCatalogProvisionedProductConfig_step2_params(saltedName string, paramValue int) string {
	return composeConfig(
		testAccAWSServiceCatalogProvisionedProductConfig_step1(saltedName),
		testAccAWSServiceCatalogProvisionedProductConfig_provider(),
		fmt.Sprintf(`

resource "aws_s3_bucket_object" "template_params" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev2_params.yaml"
  content = <<EOF
AWSTemplateFormatVersion: "2010-09-09"
Description: "Test CF template for Service Catalog terraform dev - with params"
Parameters:
    NumberWithRange:
        Type: Number
        MinValue: 1
        MaxValue: 100
        Default: 50
        Description: Enter a number between 1 and 100, default is 50
    
    Secret:
        NoEcho: true
        Description: A secret value
        Type: String
        MinLength: 1
        MaxLength: 16
        AllowedPattern: '[a-zA-Z][a-zA-Z0-9]*'
        ConstraintDescription: must begin with a letter and contain only alphanumeric characters.
Resources:
    Empty:
        Type: AWS::CloudFormation::WaitConditionHandle
Outputs:
    NumberWithRange:
        Value: !Ref NumberWithRange

EOF
}

resource "aws_servicecatalog_product" "test_params" {
  name                = %[1]q
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact {
    description = "arbitrary description"
    name        = %[1]q
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template_params.key}"
    }
  }
}

resource "aws_servicecatalog_portfolio_product_association" "test_params" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test_params.id
}

resource "aws_servicecatalog_provisioned_product" "test_params" {
    provider = aws.product-allowed-role
    provisioned_product_name = %[1]q
    product_id               = aws_servicecatalog_product.test_params.id
    provisioning_artifact_id = aws_servicecatalog_product.test_params.provisioning_artifact[0].id
    provisioning_parameters = {
        NumberWithRange = %[2]d
        Secret = "s3cr3t"
    }
    depends_on = [
      aws_iam_role_policy.test,
      aws_servicecatalog_portfolio_product_association.test_params,
      aws_servicecatalog_portfolio_principal_association.test,
    ]
}
`, saltedName, paramValue))
}
