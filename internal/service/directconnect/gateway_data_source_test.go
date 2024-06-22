// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway.test"
	datasourceName := "data.aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGatewayDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`Direct Connect Gateway not found`),
			},
			{
				Config: testAccGatewayDataSourceConfig_name(rName, sdkacctest.RandIntRange(64512, 65534)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "amazon_side_asn", resourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
				),
			},
		},
	})
}

func testAccGatewayDataSourceConfig_name(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "wrong" {
  amazon_side_asn = "%d"
  name            = "%s-wrong"
}

resource "aws_dx_gateway" "test" {
  amazon_side_asn = "%d"
  name            = "%s"
}

data "aws_dx_gateway" "test" {
  name = aws_dx_gateway.test.name
}
`, rBgpAsn+1, rName, rBgpAsn, rName)
}

const testAccGatewayDataSourceConfig_nonExistent = `
data "aws_dx_gateway" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
