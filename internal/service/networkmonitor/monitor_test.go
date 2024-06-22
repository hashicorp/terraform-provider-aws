// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmonitor/types"
	"github.com/aws/aws-sdk-go/service/networkmonitor"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
		ErrorCheck:               acctest.ErrorCheck(t, networkmonitor.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccNetworkMonitorMonitor_updates(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmonitor.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "30"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				Config: testAccMonitorConfig_tags(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation_period", "60"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
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
		ErrorCheck:               acctest.ErrorCheck(t, networkmonitor.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkmonitor.ResourceMonitor, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			var nfe *awstypes.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}

			if err != nil {
				return err
			}

			return create.Error(names.NetworkMonitor, create.ErrActionCheckingDestroyed, tfnetworkmonitor.ResNameMonitor, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMonitorExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameMonitor, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameMonitor, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkMonitorClient(ctx)

		_, err := tfnetworkmonitor.FindMonitorByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameMonitor, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccMonitorConfig(rName string, aggregation int) string {
	return fmt.Sprintf(`
resource "aws_networkmonitor_monitor" "test" {
  aggregation_period = %[2]d
  monitor_name       = %[1]q
  tags = {
    tag1 = %[1]q
  }
}
`, rName, aggregation)
}

func testAccMonitorConfig_tags(rName string, aggregation int) string {
	return fmt.Sprintf(`
resource "aws_networkmonitor_monitor" "test" {
  aggregation_period = %[2]d
  monitor_name       = %[1]q
  tags = {
    tag1 = %[1]q
    tag2 = %[1]q
  }
}
`, rName, aggregation)
}
