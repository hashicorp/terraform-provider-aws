package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAWSServiceCatalogPortfolioProductAssociation_Basic(t *testing.T) {
    salt := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogPortfolioProductAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic(salt),
				Check: testAccCheckAwsServiceCatalogPortfolioProductAssociation(),
			},
			{
				ResourceName: "aws_servicecatalog_portfolio_product_association.association",
				ImportState: true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckServiceCatalogPortfolioProductAssociationDestroy(s *terraform.State) error {
    conn := testAccProvider.Meta().(*AWSClient).scconn
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "aws_servicecatalog_portfolio_product_association" {
            continue // not our monkey
        }
        _, productId, portfolioId := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
        input := servicecatalog.ListPortfoliosForProductInput{ProductId: &productId}
        portfolios, err := conn.ListPortfoliosForProduct(&input)
        if err != nil {
            if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
                return nil // not found for product is good
            }
            return err // some other unexpected error
        }
        for _, portfolioDetail := range portfolios.PortfolioDetails {
            if *portfolioDetail.Id == portfolioId {
                return fmt.Errorf("expected AWS Service Catalog Portfolio Product Association to be gone, but it was still found")
            }
        }
    }
    return nil
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociation() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_portfolio_product_association" {
				continue // not our monkey
			}
			_, portfolioId, productId := parseServiceCatalogPortfolioProductAssociationResourceId(rs.Primary.ID)
			input := servicecatalog.ListPortfoliosForProductInput{ProductId: &productId}
			portfolios, err := conn.ListPortfoliosForProduct(&input)
			if err != nil {
				return err
			}
			for _, portfolioDetail := range portfolios.PortfolioDetails {
				if *portfolioDetail.Id == portfolioId {
					return nil //is good
				}
			}
			return fmt.Errorf("association not found between portfolio %s and product %s", portfolioId, productId)
		}
		return fmt.Errorf("no associations found")
	}
}

func testAccCheckAwsServiceCatalogPortfolioProductAssociationConfigBasic(salt string) string {
    portfolio_cfg := testAccCheckAwsServiceCatalogPortfolioResourceConfigBasic("tfm_automated_test");
    
    arbitraryBucketName := fmt.Sprintf("bucket-%s", salt)
    arbitraryProductName := fmt.Sprintf("product-%s", salt)
    arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", salt)
    p_tag1 := "FooKey = \"bar\""
    p_tag2 := "BarKey = \"foo\""
    product_cfg := testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, p_tag1, p_tag2)

    return portfolio_cfg + "\n" + product_cfg + "\n" + fmt.Sprintf(`
resource "aws_servicecatalog_portfolio_product_association" "association" {
    portfolio_id = aws_servicecatalog_portfolio.test.id
    product_id = aws_servicecatalog_product.test.id
}`)
}
