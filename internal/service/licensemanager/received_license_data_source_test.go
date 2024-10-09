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

func TestAccLicenseManagerReceivedLicenseDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_licensemanager_received_license.test"
	licenseARN := envvar.SkipIfEmpty(t, licenseARNKey, envVarLicenseARNKeyError)
	homeRegion := envvar.SkipIfEmpty(t, homeRegionKey, envVarHomeRegionError)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReceivedLicenseDataSourceConfig_arn(licenseARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGlobalARN(datasourceName, "beneficiary", "iam", "root"),
					resource.TestCheckResourceAttr(datasourceName, "consumption_configuration.#", acctest.Ct1),
					acctest.CheckResourceAttrRFC3339(datasourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(datasourceName, "entitlements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "home_region", homeRegion),
					resource.TestCheckResourceAttr(datasourceName, "issuer.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "license_arn", licenseARN),
					resource.TestCheckResourceAttrSet(datasourceName, "license_metadata.0.%"),
					resource.TestCheckResourceAttrSet(datasourceName, "license_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "product_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "product_sku"),
					resource.TestCheckResourceAttr(datasourceName, "received_metadata.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(datasourceName, "validity.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrVersion),
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
