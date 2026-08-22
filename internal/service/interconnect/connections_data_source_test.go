// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInterconnectConnectionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_connections.test"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConnections(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_basic(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.bandwidth"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.environment_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.interconnect_provider"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.location"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.state"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.type"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, region),
				),
			},
		},
	})
}

// TestAccInterconnectConnectionsDataSource_filtered exercises the environment_id and
// state filters using values taken from an existing Connection, so the filtered result
// is guaranteed to match at least that Connection whatever state it happens to be in.
func TestAccInterconnectConnectionsDataSource_filtered(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_connections.test"
	allDataSourceName := "data.aws_interconnect_connections.all"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConnections(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_filtered(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment_id", allDataSourceName, "connections.0.environment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, allDataSourceName, "connections.0.state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "connections.0.state", allDataSourceName, "connections.0.state"),
				),
			},
		},
	})
}

// TestAccInterconnectConnectionsDataSource_providerFilter exercises the
// interconnect_provider filter, taking the provider name from an existing Connection so
// the filtered result matches at least that Connection.
func TestAccInterconnectConnectionsDataSource_providerFilter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_connections.test"
	allDataSourceName := "data.aws_interconnect_connections.all"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConnections(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_providerFilter(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.0.id"),
					resource.TestCheckResourceAttr(dataSourceName, "interconnect_provider.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "interconnect_provider.0.cloud_service_provider", allDataSourceName, "connections.0.interconnect_provider"),
					resource.TestCheckResourceAttrPair(dataSourceName, "connections.0.interconnect_provider", allDataSourceName, "connections.0.interconnect_provider"),
				),
			},
		},
	})
}

// testAccPreCheckConnections skips the test unless the given Region has at least one
// Interconnect Connection, since the Connection data sources have nothing to read
// otherwise and Connections cannot be created without an out-of-band confirmation on
// the partner's side.
func testAccPreCheckConnections(ctx context.Context, t *testing.T, region string) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

	input := interconnect.ListConnectionsInput{}
	output, err := conn.ListConnections(ctx, &input, func(o *interconnect.Options) {
		o.Region = region
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output.Connections) == 0 {
		t.Skipf("skipping acceptance testing: no Interconnect Connections in %s, set %s to a Region that has them", region, envVarRegion)
	}
}

func testAccConnectionsDataSourceConfig_basic(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_connections" "test" {
  region = %[1]q
}
`, region)
}

func testAccConnectionsDataSourceConfig_filtered(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_connections" "all" {
  region = %[1]q
}

data "aws_interconnect_connections" "test" {
  region         = %[1]q
  environment_id = data.aws_interconnect_connections.all.connections[0].environment_id
  state          = data.aws_interconnect_connections.all.connections[0].state
}
`, region)
}

func testAccConnectionsDataSourceConfig_providerFilter(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_connections" "all" {
  region = %[1]q
}

data "aws_interconnect_connections" "test" {
  region = %[1]q

  interconnect_provider {
    cloud_service_provider = data.aws_interconnect_connections.all.connections[0].interconnect_provider
  }
}
`, region)
}
