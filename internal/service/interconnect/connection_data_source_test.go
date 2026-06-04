// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInterconnectConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_interconnect_connection.test"
	resourceName := "aws_interconnect_connection.test"
	environmentID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ENVIRONMENT_ID")
	directConnectGatewayID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID")
	remoteAccountIdentifier := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_REMOTE_ACCOUNT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(environmentID, directConnectGatewayID, remoteAccountIdentifier),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment_id", resourceName, "environment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(environmentID, directConnectGatewayID, remoteAccountIdentifier string) string {
	return fmt.Sprintf(`
resource "aws_interconnect_connection" "test" {
  bandwidth      = "1Gbps"
  environment_id = %[1]q

  attach_point {
    direct_connect_gateway = %[2]q
  }

  remote_account {
    identifier = %[3]q
  }
}

data "aws_interconnect_connection" "test" {
  id = aws_interconnect_connection.test.id
}
`, environmentID, directConnectGatewayID, remoteAccountIdentifier)
}
