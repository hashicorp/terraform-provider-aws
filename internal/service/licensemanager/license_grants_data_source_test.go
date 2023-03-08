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

func TestAccLicenseManagerGrantsDataSource_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"grant": {
			"basic": testAccGrantsDataSource_basic,
			"empty": testAccGrantsDataSource_empty,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccGrantsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_licensemanager_grants.test"
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	licenseARN := os.Getenv(licenseKey)
	if licenseARN == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantsDataSourceConfig_arns(licenseARN, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "arns.0"),
				),
			},
		},
	})
}

func testAccGrantsDataSource_empty(t *testing.T) {
	datasourceName := "data.aws_licensemanager_grants.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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
  license_arn = %[1]q
}

locals {
  allowed_operations = [for i in data.aws_licensemanager_received_license.test.received_metadata[0].allowed_operations : i if i != "CreateGrant"]
}

data "aws_partition" "current" {
  provider = awsalternate
}
data "aws_caller_identity" "current" {
  provider = awsalternate
}

resource "aws_licensemanager_grant" "test" {
  name               = %[2]q
  allowed_operations = local.allowed_operations
  license_arn        = data.aws_licensemanager_received_license.test.license_arn
  principal          = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
}

data "aws_licensemanager_grants" "test" {}
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
