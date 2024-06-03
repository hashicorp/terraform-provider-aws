// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service.test"
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_name", dataSourceName, "custom_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_entry.#", dataSourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccVPCLatticeServiceDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service.test"
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_name", dataSourceName, "custom_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_entry.#", dataSourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "service_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccVPCLatticeServiceDataSource_shared(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service.test"
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_shared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_name", dataSourceName, "custom_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_entry.#", dataSourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckNoResourceAttr(dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_vpclattice_service" "test" {
  service_identifier = aws_vpclattice_service.test.id
}
`, rName)
}

func testAccServiceDataSourceConfig_byName(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_vpclattice_service" "test" {
  name = aws_vpclattice_service.test.name
}
`, rName)
}

func testAccServiceDataSourceConfig_shared(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "source" {}

data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_vpclattice_service" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = false
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_vpclattice_service.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.target.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_vpclattice_service" "test" {
  provider = "awsalternate"

  service_identifier = aws_vpclattice_service.test.id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName))
}
