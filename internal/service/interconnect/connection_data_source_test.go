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

func TestAccInterconnectConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_connection.test"
	connectionsDataSourceName := "data.aws_interconnect_connections.test"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConnections(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Compare against the plural data source for the same Connection, so a
					// flattening error in either would be caught.
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, connectionsDataSourceName, "connections.0.id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, connectionsDataSourceName, "connections.0.arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bandwidth", connectionsDataSourceName, "connections.0.bandwidth"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment_id", connectionsDataSourceName, "connections.0.environment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "interconnect_provider", connectionsDataSourceName, "connections.0.interconnect_provider"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrLocation, connectionsDataSourceName, "connections.0.location"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shared_id", connectionsDataSourceName, "connections.0.shared_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, connectionsDataSourceName, "connections.0.state"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, connectionsDataSourceName, "connections.0.type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attach_point.#", connectionsDataSourceName, "connections.0.attach_point.#"),
					// Members returned only by GetConnection, and absent from ConnectionSummary.
					resource.TestCheckResourceAttrSet(dataSourceName, "activation_key"),
					resource.TestCheckResourceAttrSet(dataSourceName, "owner_account"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.%"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, region),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_connections" "test" {
  region = %[1]q
}

data "aws_interconnect_connection" "test" {
  region = %[1]q
  id     = data.aws_interconnect_connections.test.connections[0].id
}
`, region)
}
