package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSServiceCatalogConstraintDataSource_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_constraint.test"
	dataSourceName := "data.aws_servicecatalog_constraint.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogConstraintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogConstraintDataSourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogConstraintExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "accept_language", dataSourceName, "accept_language"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "owner", dataSourceName, "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "parameters", dataSourceName, "parameters"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", dataSourceName, "portfolio_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", dataSourceName, "product_id"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
				),
			},
		},
	})
}

func testAccAWSServiceCatalogConstraintDataSourceConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccAWSServiceCatalogConstraintConfig_basic(rName, description), `
data "aws_servicecatalog_constraint" "test" {
  id = aws_servicecatalog_constraint.test.id
}
`)
}
