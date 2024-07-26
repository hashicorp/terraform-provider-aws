// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "aws_device", resourceName, "aws_device"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrLocation, resourceName, names.AttrLocation),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(datasourceName, "partner_name", resourceName, "partner_name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrProviderName, resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttrPair(datasourceName, "vlan_id", resourceName, "vlan_id"),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}

data "aws_dx_connection" "test" {
  name = aws_dx_connection.test.name
}
`, rName)
}
