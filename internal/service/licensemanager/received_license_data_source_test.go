package licensemanager_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLicenseManagerReceivedLicenseDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_licensemanager_received_license.test"
	licenseARNKey := "LICENSE_MANAGER_LICENSE_ARN"
	licenseARN := os.Getenv(licenseARNKey)
	if licenseARN == "" {
		t.Skipf("Environment variable %s is not set", licenseARNKey)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicenseDataSourceConfig_arn(licenseARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "beneficiary"),
					resource.TestCheckResourceAttr(datasourceName, "consumption_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "create_time"),
					resource.TestCheckResourceAttr(datasourceName, "entitlements.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "home_region"),
					resource.TestCheckResourceAttr(datasourceName, "issuer.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "license_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "license_metadata.0.%"),
					resource.TestCheckResourceAttrSet(datasourceName, "license_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "product_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "product_sku"),
					resource.TestCheckResourceAttr(datasourceName, "received_metadata.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "status"),
					resource.TestCheckResourceAttr(datasourceName, "validity.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "version"),
				),
			},
		},
	})
}

func testAccReceivedLicenseDataSourceConfig_arn(licenseARN string) string {
	return fmt.Sprintf(`
data "aws_licensemanager_received_license" "test" {
  license_arn = %[1]q
}
`, licenseARN)
}
