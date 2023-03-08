package licensemanager_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLicenseManagerGrantsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_licensemanager_grants.test"
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	licenseARN := os.Getenv(licenseKey)
	if licenseARN == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	resourceName := "aws_licensemanager_grant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantsDataSourceConfig_arns(licenseARN, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "arns.0", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccLicenseManagerGrantsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_licensemanager_grants.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantsDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccGrantsDataSourceConfig_arns(licenseARN string, rName string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_licensemanager_received_license" "test" {
  provider    = awsalternate
  license_arn = %[1]q
}

locals {
  allowed_operations = [for i in data.aws_licensemanager_received_license.test.received_metadata[0].allowed_operations : i if i != "CreateGrant"]
}
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_licensemanager_grant" "test" {
  provider = awsalternate

  name               = %[2]q
  allowed_operations = local.allowed_operations
  license_arn        = data.aws_licensemanager_received_license.test.license_arn
  principal          = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
}

data "aws_licensemanager_grants" "test" {
  filter {
    name = "ProductSKU"
    values = [
		%[1]q
    ]
  }
}
`, licenseARN, rName)
}

func testAccGrantsDataSourceConfig_empty() string {
	return `
data "aws_licensemanager_grants" "test" {
  filter {
    name = "LicenseIssuerName"
    values = [
      "This Is Fake"
    ]
  }
}
`
}
