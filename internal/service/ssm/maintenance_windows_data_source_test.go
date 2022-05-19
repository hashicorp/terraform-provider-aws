package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMMaintenanceWindowsDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ssm_maintenance_windows.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMaintenanceWindowsDataSourceConfig_filter_name(rName1, rName2, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", "aws_ssm_maintenance_window.test2", "id"),
				),
			},
			{
				Config: testAccCheckMaintenanceWindowsDataSourceConfig_filter_enabled(rName1, rName2, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckMaintenanceWindowsDataSourceConfig(rName1, rName2, rName3 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test1" {
  name     = "%[1]s"
  duration = 1
  cutoff   = 0
  schedule = "cron(0 16 ? * TUE *)"
}

resource "aws_ssm_maintenance_window" "test2" {
  name     = "%[2]s"
  duration = 1
  cutoff   = 0
  schedule = "cron(0 16 ? * WED *)"
}

resource "aws_ssm_maintenance_window" "test3" {
  name     = "%[3]s"
  duration = 1
  cutoff   = 0
  schedule = "cron(0 16 ? * THU *)"

  enabled = false
}
`, rName1, rName2, rName3)
}

func testAccCheckMaintenanceWindowsDataSourceConfig_filter_name(rName1, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccCheckMaintenanceWindowsDataSourceConfig(rName1, rName2, rName3),
		fmt.Sprintf(`
data "aws_ssm_maintenance_windows" "test" {
  filter {
    name   = "Name"
    values = ["%[1]s"]
  }

  depends_on = [
    aws_ssm_maintenance_window.test1,
    aws_ssm_maintenance_window.test2,
    aws_ssm_maintenance_window.test3,
  ]
}
`, rName2))
}

func testAccCheckMaintenanceWindowsDataSourceConfig_filter_enabled(rName1, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccCheckMaintenanceWindowsDataSourceConfig(rName1, rName2, rName3),
		`
data "aws_ssm_maintenance_windows" "test" {
  filter {
    name   = "Enabled"
    values = ["true"]
  }

  depends_on = [
    aws_ssm_maintenance_window.test1,
    aws_ssm_maintenance_window.test2,
    aws_ssm_maintenance_window.test3,
  ]
}
`)
}
