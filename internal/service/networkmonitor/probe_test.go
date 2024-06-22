// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmonitor/types"
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

func TestAccNetworkMonitorProbe_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_probe.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProbeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProbeConfig_basic(rName, "10.0.0.1", 8080, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProbeExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "probe.destination", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "probe.destination_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "probe.packet_size", "200"),
					resource.TestCheckResourceAttr(resourceName, "probe.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkMonitorProbe_updates(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_probe.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProbeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProbeConfig_basic(rName, "10.0.0.1", 8080, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProbeExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "probe.destination", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "probe.destination_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "probe.packet_size", "200"),
					resource.TestCheckResourceAttr(resourceName, "probe.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", rName),
				),
			},
			{
				Config: testAccProbeConfig_2tags(rName, "10.0.0.2", 8081, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProbeExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "probe.destination", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "probe.destination_port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "probe.packet_size", "300"),
					resource.TestCheckResourceAttr(resourceName, "probe.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", rName),
				),
			},
		},
	})
}

func TestAccNetworkMonitorProbe_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmonitor_probe.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "networkMonitor"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProbeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProbeConfig_basic(rName, "10.0.0.1", 8080, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProbeExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkmonitor.ResourceProbe, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProbeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkMonitorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmonitor_probe" {
				continue
			}

			_, err := tfnetworkmonitor.FindProbeByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *awstypes.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}

				if tfresource.NotFound(err) {
					return nil
				}

				return err
			}

			return create.Error(names.NetworkMonitor, create.ErrActionCheckingDestroyed, tfnetworkmonitor.ResNameMonitor, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckProbeExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameProbe, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameProbe, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkMonitorClient(ctx)

		_, err := tfnetworkmonitor.FindProbeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.NetworkMonitor, create.ErrActionCheckingExistence, tfnetworkmonitor.ResNameProbe, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccProbeConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmonitor_monitor" "test" {
  aggregation_period = 30
  monitor_name       = %[1]q
  tags = {
    tag1 = %[1]q
  }
}
`, rName)
}

func testAccProbeConfig_basic(rName, destination string, port, packetSize int) string {
	return acctest.ConfigCompose(testAccProbeConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmonitor_probe" "test" {
  monitor_name = aws_networkmonitor_monitor.test.monitor_name
  probe {
    destination      = %[2]q
    destination_port = %[3]d
    protocol         = "TCP"
    source_arn       = aws_subnet.test.arn
    packet_size      = %[4]d
  }
  tags = {
    tag1 = %[1]q
  }
}
`, rName, destination, port, packetSize))
}

func testAccProbeConfig_2tags(rName, destination string, port, packetSize int) string {
	return acctest.ConfigCompose(testAccProbeConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmonitor_probe" "test" {
  monitor_name = aws_networkmonitor_monitor.test.monitor_name
  probe {
    destination      = %[2]q
    destination_port = %[3]d
    protocol         = "TCP"
    source_arn       = aws_subnet.test.arn
    packet_size      = %[4]d
  }
  tags = {
    tag1 = %[1]q
    tag2 = %[1]q
  }
}
`, rName, destination, port, packetSize))
}
