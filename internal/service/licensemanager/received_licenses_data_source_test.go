package licensemanager_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLicenseManagerReceivedLicensesDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_licensemanager_received_licenses.test"
	productSKUKey := "LICENSE_MANAGER_LICENSE_PRODUCT_SKU"
	productSKU := os.Getenv(productSKUKey)
	if productSKU == "" {
		t.Skipf("Environment variable %s is not set", productSKUKey)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicensesDataSourceConfig_arns(productSKU),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", "1"),
				),
			},
		},
	})
}

func TestAccLicenseManagerReceivedLicensesDataSource_empty(t *testing.T) {
	datasourceName := "data.aws_licensemanager_received_licenses.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicensesDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccReceivedLicensesDataSourceConfig_arns(productSKU string) string {
	return fmt.Sprintf(`
data "aws_licensemanager_received_licenses" "test" {
  filter {
    name = "ProductSKU"
    values = [
		%[1]q
    ]
  }
}
`, productSKU)
}

func testAccReceivedLicensesDataSourceConfig_empty() string {
	return `
data "aws_licensemanager_received_licenses" "test" {
  filter {
    name = "IssuerName"
    values = [
      "This Is Fake"
    ]
  }
}
`
}
