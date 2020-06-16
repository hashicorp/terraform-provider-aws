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
				Config: testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProduct(resourceName),
					// TODO addl checks (ones below are wrong)
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

func testAccCheckAwsServiceCatalogProvisionedProductResourceConfigTemplate(provisionedProductName string) string {

    return fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
    provisioned_product_name = "%s"
    product_id = "prod-lgqvxr6phzrzk"
    provisioning_artifact_id = "pa-5bddiphhjdsoy"
}
`, provisionedProductName)

    // TODO test should create the portfolio etc as below, but we need portfolio_product_association for that
    
/*

    portfolio_cfg := testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic("tfm_automated_test_portfolio");
    
    arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
    arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
    arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
    p_tag1 := "FooKey = \"bar\""
    p_tag2 := "BarKey = \"foo\""
    product_cfg := testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, p_tag1, p_tag2)

    return portfolio_cfg + "\n" + product_cfg + "\n" + fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_product_association" "association" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}

resource "aws_servicecatalog_provisioned_product" "test" {
    provisioned_product_name = "%s"
    product_id = aws_servicecatalog_product.test.id
    provisioning_artifact_id = aws_servicecatalog_product.test.provisioning_artifact[0].id
}
`, provisionedProductName)
*/
}
