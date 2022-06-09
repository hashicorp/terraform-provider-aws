package ce_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccCEAnomalyMonitor_Dimensional_serial ensures all tests for monitor_type DIMENSIONAL run in series
// before any tests are run in parallel. This is required as AWS only allows 1 Anomaly Monitor
// with a type of DIMENSIONAL per AWS account.
func TestAccCEAnomalyMonitor_Dimensional_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AnomalyMonitor": {
			"basic":      testAccAnomalyMonitor_basic,
			"disappears": testAccAnomalyMonitor_disappears,
			"name":       testAccAnomalyMonitor_Name,
			"tags":       testAccAnomalyMonitor_Tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAnomalyMonitor_basic(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ce", regexp.MustCompile(`anomalymonitor/.+`)),
					resource.TestCheckResourceAttr(resourceName, "monitor_dimension", "SERVICE"),
					resource.TestCheckResourceAttr(resourceName, "monitor_type", "DIMENSIONAL"),
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

func testAccAnomalyMonitor_Name(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalyMonitorConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccCEAnomalyMonitor_Custom(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_Custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccAnomalyMonitor_Tags(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalyMonitorConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAnomalyMonitorConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAnomalyMonitor_disappears(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfce.ResourceAnomalyMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAnomalyMonitorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cost Explorer Anomaly Monitor is set")
		}

		resp, err := conn.GetAnomalyMonitors(&costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return fmt.Errorf("Error describing Cost Explorer Anomaly Monitor: %s", err.Error())
		}

		if resp == nil || len(resp.AnomalyMonitors) < 1 {
			return fmt.Errorf("Anomaly Monitor (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckAnomalyMonitorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_monitor" {
			continue
		}

		resp, err := conn.GetAnomalyMonitors(&costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalyMonitor, rs.Primary.ID, err)
		}

		if resp != nil && len(resp.AnomalyMonitors) > 0 {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalyMonitor, rs.Primary.ID, errors.New("still exists"))
		}
	}

	return nil

}

func testAccAnomalyMonitorConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name              = %[1]q
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}
`, rName)
}

func testAccAnomalyMonitorConfig_Custom(rName string) string {
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
		"Key": "CostCenter",
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

func testAccAnomalyMonitorConfig_Tags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`	
resource "aws_ce_anomaly_monitor" "test" {
  name              = %[1]q
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAnomalyMonitorConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`	
resource "aws_ce_anomaly_monitor" "test" {
  name              = %[1]q
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
