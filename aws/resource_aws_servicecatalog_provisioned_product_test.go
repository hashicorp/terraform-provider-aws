package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"testing"
)

func TestAccAWSServiceCatalogProvisionedProduct_Basic(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioned_product.test"
	salt := acctest.RandString(5)
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// need multiple independent providers for assume-role not to leak
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1(salt),
			},
			{
				// provisioning with assume_role needs to be run in a second step
				// because a provider can only assume a role existing before its definition
				// see https://github.com/hashicorp/terraform/issues/2430
				Config: testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate2(salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProduct(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalog", regexp.MustCompile(`stack/.+/pp-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", "tfm-sc-test-pp-"+salt),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckProvisionedProduct(pr string) resource.TestCheckFunc {
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

		_, err := conn.DescribeProvisionedProduct(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceCatalogProvisionedProductDestroy(s *terraform.State) error {
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

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplatePolicy() string {
	return fmt.Sprintf(`

resource "aws_iam_role_policy" "test_sc" {
  name = "test_policy"
  role = aws_iam_role.tfm-sc-tester.id

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

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1(salt string) string {
	return testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigRole(salt) +
		testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplatePolicy() +
		testAccCheckAwsServiceCatalogPortfolioPrincipalAssociationConfigAssociation() +
		testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic(salt)
}

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate2(salt string) string {
	provisionedProductName := "tfm-sc-test-pp-" + salt
	x := testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1(salt) +
		fmt.Sprintf(`

provider "aws" {
  alias               = "product-allowed-role"
  assume_role {
    role_arn          = aws_iam_role.tfm-sc-tester.arn
    session_name      = "tfm-sc-testing"
    external_id       = "tfm-sc-testing"
  }
}

`) + fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
    provider = aws.product-allowed-role
    provisioned_product_name = "%s"
    product_id               = aws_servicecatalog_product.test.id
    provisioning_artifact_id = aws_servicecatalog_product.test.provisioning_artifact[0].id
    depends_on = [
      aws_iam_role_policy.test_sc,
      aws_servicecatalog_portfolio_product_association.association,
      aws_servicecatalog_portfolio_principal_association.association,
    ]
}`, provisionedProductName)
	log.Printf("[DEBUG] provision-product test using configuration:\n" + x)
	return x
}
