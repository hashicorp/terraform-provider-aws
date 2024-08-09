// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccGrantsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_licensemanager_grants.test"
	licenseARN := envvar.SkipIfEmpty(t, licenseARNKey, envVarLicenseARNKeyError)
	principal := envvar.SkipIfEmpty(t, principalKey, envVarPrincipalKeyError)
	homeRegion := envvar.SkipIfEmpty(t, homeRegionKey, envVarHomeRegionError)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantsDataSourceConfig_basic(licenseARN, rName, principal, homeRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "arns.0"),
				),
			},
		},
	})
}

func testAccGrantsDataSource_noMatch(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_licensemanager_grants.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantsDataSourceConfig_noMatch(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccGrantsDataSourceConfig_basic(licenseARN, rName, principal, homeRegion string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRegionalProvider(homeRegion),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_licensemanager_received_license" "test" {
  license_arn = %[1]q
}

locals {
  allowed_operations = [for i in data.aws_licensemanager_received_license.test.received_metadata[0].allowed_operations : i if i != "CreateGrant"]
}

resource "aws_licensemanager_grant" "test" {
  name               = %[2]q
  allowed_operations = local.allowed_operations
  license_arn        = data.aws_licensemanager_received_license.test.license_arn
  principal          = %[3]q
}

data "aws_licensemanager_grants" "test" {}
`, licenseARN, rName, principal),
	)
}

func testAccGrantsDataSourceConfig_noMatch() string {
	return `
data "aws_licensemanager_grants" "test" {
  filter {
    name = "LicenseIssuerName"
    values = [
      "No Match"
    ]
  }
}
`
}
