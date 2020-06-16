package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"testing"
)

func TestAccAWSServiceCatalogProvisionedProduct_Basic(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioned_product.test"
	name := acctest.RandString(5)
	
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
		    {
		        Config: testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1(),
		    },
			{
				Config: testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProduct(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalog", regexp.MustCompile(`stack/.+/pp-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_product_name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate3(provisionedProductName string) string {
    return fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {    
    provisioned_product_name = "%s"
    product_id = "prod-lgqvxr6phzrzk"
    provisioning_artifact_id = "pa-5bddiphhjdsoy"
}`, provisionedProductName) 
}

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1() string {
    return testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic() + "\n" +
        testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigRoleAndAssociation() + "\n" + 
        fmt.Sprintf(`

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
          "cloudformation:CreateStack"
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

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate2(provisionedProductName string) string {
    // needs to be run in a second step due to https://github.com/hashicorp/terraform/issues/2430
    // (we can't have "depends_on" in a provider)
    return testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate1() + fmt.Sprintf(`
    
provider "aws" {
  alias          = "product-allowed-role"
  assume_role {
    role_arn          = aws_iam_role.tfm-sc-tester.arn
    session_name      = "tfm-sc-testing"
  }
}

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
}
`, provisionedProductName)
}
