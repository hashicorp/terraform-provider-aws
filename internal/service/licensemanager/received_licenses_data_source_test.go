// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLicenseManagerReceivedLicensesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_licensemanager_received_licenses.test"
	licenseARN := envvar.SkipIfEmpty(t, licenseARNKey, envVarLicenseARNKeyError)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicensesDataSourceConfig_arns(licenseARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccLicenseManagerReceivedLicensesDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_licensemanager_received_licenses.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicensesDataSourceConfig_empty(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccReceivedLicensesDataSourceConfig_arns(licenseARN string) string {
	return fmt.Sprintf(`
data "aws_licensemanager_received_licenses" "test" {
  filter {
    name = "ProductSKU"
    values = [
      data.aws_licensemanager_received_license.test.product_sku
    ]
  }
}

data "aws_licensemanager_received_license" "test" {
  license_arn = %[1]q
}
`, licenseARN)
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
