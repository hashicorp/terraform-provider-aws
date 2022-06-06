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

// TestAccAnomalyMonitor_dimensionalserial limits the number of parallel tests run with a type of DIMENSIONAL to 1.
// This is required as AWS only allows 1 Anomaly Monitor with a type of DIMENSIONAL per AWS account.
func TestAccAnomalyMonitor_dimensionalserial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContainerService": {
			"basic":      testAccAnomalyMonitor_basic,
			"disappears": testAccAnomalyMonitor_disappears,
			"name":       testAccAnomalyMonitor_Name,
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
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dimensionValue := "SERVICE"
	dimensionBadValue := "BADVALUE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config:      testAccAnomalyMonitorConfig(rName, dimensionBadValue),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected dimension to be one of \[SERVICE\], got %s`, dimensionBadValue)),
			},
			{
				Config: testAccAnomalyMonitorConfig(rName, dimensionValue),
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
		},
	})
}

func testAccAnomalyMonitor_Name(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	dimensionValue := "SERVICE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig(rName, dimensionValue),
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
				Config: testAccAnomalyMonitorConfig(rName2, dimensionValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAnomalyMonitor_Custom(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

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

func testAccAnomalyMonitor_disappears(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyMonitorConfig(rName, "SERVICE"),
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
			return errors.New("No Lightsail Database ID is set")
		}

		resp, err := conn.GetAnomalyMonitors(&costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return err
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

func testAccAnomalyMonitorConfig(rName string, dimension string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name      = %[1]q
  type      = "DIMENSIONAL"
  dimension = %[2]q
}
`, rName, dimension)
}

func testAccAnomalyMonitorConfig_Custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
