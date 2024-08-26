// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCEAnomalyMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var monitor awstypes.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyMonitorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ce", regexache.MustCompile(`anomalymonitor/.+`)),
					resource.TestCheckResourceAttr(resourceName, "monitor_type", "CUSTOM"),
					resource.TestCheckResourceAttrSet(resourceName, "monitor_specification"),
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

func TestAccCEAnomalyMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var monitor awstypes.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyMonitorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfce.ResourceAnomalyMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCEAnomalyMonitor_update(t *testing.T) {
	ctx := acctest.Context(t)
	var monitor awstypes.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyMonitorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalyMonitorConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccCEAnomalyMonitor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var monitor awstypes.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyMonitorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalyMonitorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAnomalyMonitorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// An AWS account can only have one anomaly monitor of type DIMENSIONAL. As
// such, if additional tests are added, they should be combined with the
// following test in a serial test
func TestAccCEAnomalyMonitor_Dimensional(t *testing.T) {
	ctx := acctest.Context(t)
	var monitor awstypes.AnomalyMonitor
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyMonitorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_dimensional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(ctx, resourceName, &monitor),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "monitor_type", "DIMENSIONAL"),
					resource.TestCheckResourceAttr(resourceName, "monitor_dimension", "SERVICE"),
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

func testAccCheckAnomalyMonitorExists(ctx context.Context, n string, v *awstypes.AnomalyMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		output, err := tfce.FindAnomalyMonitorByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAnomalyMonitorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ce_anomaly_monitor" {
				continue
			}

			_, err := tfce.FindAnomalyMonitorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cost Explorer Anomaly Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAnomalyMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name         = %[1]q
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "user:CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON
}
`, rName)
}

func testAccAnomalyMonitorConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`	
resource "aws_ce_anomaly_monitor" "test" {
  name         = %[1]q
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "user:CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAnomalyMonitorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`	
resource "aws_ce_anomaly_monitor" "test" {
  name         = %[1]q
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "user:CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAnomalyMonitorConfig_dimensional(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name              = %[1]q
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}
`, rName)
}
