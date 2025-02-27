// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmonitor "github.com/hashicorp/terraform-provider-aws/internal/service/networkmonitor"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkMonitorMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "60"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "networkmonitor", "monitor/{monitor_name}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "monitor_name"),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_aggregationPeriod(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "30"),
				),
			},
		},
	})
}

func TestAccNetworkMonitorMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkmonitor.ResourceMonitor, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkMonitorMonitor_aggregationPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_aggregationPeriod(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_aggregationPeriod(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "60"),
				),
			},
		},
	})
}

func testAccCheckMonitorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkMonitorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmonitor_monitor" {
				continue
			}

			_, err := tfnetworkmonitor.FindMonitorByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Network Monitor Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMonitorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkMonitorClient(ctx)

		_, err := tfnetworkmonitor.FindMonitorByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmonitor_monitor" "test" {
  monitor_name = %[1]q
}
`, rName)
}

func testAccMonitorConfig_aggregationPeriod(rName string, aggregation int) string {
	return fmt.Sprintf(`
resource "aws_networkmonitor_monitor" "test" {
  monitor_name       = %[1]q
  aggregation_period = %[2]d
}
`, rName, aggregation)
}
