package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccServiceCatalogProductDataSource_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "owner", dataSourceName, "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Name", dataSourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccServiceCatalogProductDataSource_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_physicalID(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "owner", dataSourceName, "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Name", dataSourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccProductDataSourceConfig_basic(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccProductConfig_basic(rName, description, supportDescription, domain, email), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}

func testAccProductDataSourceConfig_physicalID(rName, domain, email string) string {
	return acctest.ConfigCompose(testAccProductConfig_physicalID(rName, domain, email), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}
