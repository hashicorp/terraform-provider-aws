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

func TestAccVPCLatticeServiceNetworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"
	dataSourceName := "data.aws_vpclattice_service_network.test"

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
				Config: testAccServiceNetworkDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "number_of_associated_services", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "number_of_associated_vpcs", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccVPCLatticeServiceNetworkDataSource_shared(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"
	dataSourceName := "data.aws_vpclattice_service_network.test"

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
				Config: testAccServiceNetworkDataSourceConfig_shared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "number_of_associated_services", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "number_of_associated_vpcs", acctest.Ct0),
					resource.TestCheckNoResourceAttr(dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccServiceNetworkDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`  
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_vpclattice_service_network" "test" {
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName)
}

func testAccServiceNetworkDataSourceConfig_shared(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "source" {}

data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_vpclattice_service_network" "test" {
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
  resource_arn       = aws_vpclattice_service_network.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.target.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_vpclattice_service_network" "test" {
  provider = "awsalternate"

  service_network_identifier = aws_vpclattice_service_network.test.id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName))
}
