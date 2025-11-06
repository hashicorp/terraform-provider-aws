// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfnetworkflowmonitor "github.com/hashicorp/terraform-provider-aws/internal/service/networkflowmonitor"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFlowMonitorMonitor_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Monitor": {
			acctest.CtBasic:      testAccNetworkFlowMonitorMonitor_basic,
			acctest.CtDisappears: testAccNetworkFlowMonitorMonitor_disappears,
			"tags":               testAccNetworkFlowMonitorMonitor_tags,
			"update":             testAccNetworkFlowMonitorMonitor_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccNetworkFlowMonitorMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkflowmonitor_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "networkflowmonitor", fmt.Sprintf("monitor/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "monitor_status"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scope_arn"},
			},
		},
	})
}

func testAccNetworkFlowMonitorMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkflowmonitor_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkflowmonitor.ResourceMonitor, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNetworkFlowMonitorMonitor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkflowmonitor_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scope_arn"},
			},
			{
				Config: testAccMonitorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMonitorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccNetworkFlowMonitorMonitor_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkflowmonitor_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "local_resources.0.type", "AWS::EC2::VPC"),
				),
			},
			//adding one more local resource to monitor
			{
				Config: testAccMonitorConfig_updated1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "local_resources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "remote_resources.#", "1"),
				),
			},
			//reverting local resources to single resource.
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "local_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "local_resources.0.type", "AWS::EC2::VPC"),
				),
			},
			//adding 2 local resources and 2 remote resources
			{
				Config: testAccMonitorConfig_updated2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "local_resources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "remote_resources.#", "2"),
				),
			},
			//reverting local resources to single resource.
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "local_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "local_resources.0.type", "AWS::EC2::VPC"),
				),
			},
		},
	})
}

func testAccCheckMonitorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFlowMonitorClient(ctx)

		_, err := tfnetworkflowmonitor.FindMonitorByName(ctx, conn, rs.Primary.Attributes["monitor_name"])

		return err
	}
}

func testAccCheckMonitorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFlowMonitorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkflowmonitor_monitor" {
				continue
			}

			_, err := tfnetworkflowmonitor.FindMonitorByName(ctx, conn, rs.Primary.Attributes["monitor_name"])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Flow Monitor Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = %[1]q
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }
}
`, rName)
}

func testAccMonitorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = %[1]q
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMonitorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = %[1]q
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccMonitorConfig_updated1(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = %[1]q
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  local_resources {
    type       = "AWS::EC2::Subnet"
    identifier = aws_subnet.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }
}
`, rName)
}

func testAccMonitorConfig_updated2(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = %[1]q
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  local_resources {
    type       = "AWS::EC2::Subnet"
    identifier = aws_subnet.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::Subnet"
    identifier = aws_subnet.test.arn
  }
}
`, rName)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFlowMonitorClient(ctx)

	input := networkflowmonitor.ListMonitorsInput{}

	_, err := conn.ListMonitors(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
