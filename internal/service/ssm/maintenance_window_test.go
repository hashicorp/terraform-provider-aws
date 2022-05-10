package ssm_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSSMMaintenanceWindow_basic(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule_offset", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSSMMaintenanceWindow_description(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowDescriptionConfig(rName, "foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "foo"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowDescriptionConfig(rName, "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "bar"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_tags(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
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
				Config: testAccMaintenanceWindowTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMaintenanceWindowTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_disappears(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &winId),
					testAccCheckMaintenanceWindowDisappears(&winId),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_multipleUpdates(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				Config: testAccMaintenanceWindowMultipleUpdatesConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "8"),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_cutoff(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowCutoffConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowCutoffConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "2"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_duration(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowDurationConfig(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowDurationConfig(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_enabled(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowEnabledConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_endDate(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	endDate1 := time.Now().UTC().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	endDate2 := time.Now().UTC().Add(730 * 24 * time.Hour).Format(time.RFC3339)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowEndDateConfig(rName, endDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowEndDateConfig(rName, endDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate2),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_schedule(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowScheduleConfig(rName, "cron(0 16 ? * TUE *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowScheduleConfig(rName, "cron(0 16 ? * WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_scheduleTimezone(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowScheduleTimezoneConfig(rName, "America/Los_Angeles"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/Los_Angeles"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowScheduleTimezoneConfig(rName, "America/New_York"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/New_York"),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", ""),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_scheduleOffset(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowScheduleOffsetConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule_offset", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowScheduleOffsetConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule_offset", "5"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_startDate(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	startDate1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	startDate2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowStartDateConfig(rName, startDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowStartDateConfig(rName, startDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate2),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
				),
			},
		},
	})
}

func testAccCheckMaintenanceWindowExists(n string, res *ssm.MaintenanceWindowIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		resp, err := conn.DescribeMaintenanceWindows(&ssm.DescribeMaintenanceWindowsInput{
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("Name"),
					Values: []*string{aws.String(rs.Primary.Attributes["name"])},
				},
			},
		})
		if err != nil {
			return err
		}

		for _, i := range resp.WindowIdentities {
			if *i.WindowId == rs.Primary.ID {
				*res = *i
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window found")
	}
}

func testAccCheckMaintenanceWindowDisappears(maintenanceWindowIdentity *ssm.MaintenanceWindowIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		id := aws.StringValue(maintenanceWindowIdentity.WindowId)
		_, err := conn.DeleteMaintenanceWindow(&ssm.DeleteMaintenanceWindowInput{
			WindowId: aws.String(id),
		})
		if err != nil {
			return fmt.Errorf("error deleting maintenance window %s: %s", id, err)
		}
		return nil
	}
}

func testAccCheckMaintenanceWindowDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window" {
			continue
		}

		out, err := conn.GetMaintenanceWindow(&ssm.GetMaintenanceWindowInput{
			WindowId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *out.WindowId == rs.Primary.ID {
				return fmt.Errorf("SSM Maintenance Window %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the SSM Maintenance Window is already destroyed
		if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
			continue
		}

		return err
	}

	return nil
}

func testAccMaintenanceWindowConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName)
}

func testAccMaintenanceWindowDescriptionConfig(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff      = 1
  duration    = 3
  name        = %[1]q
  description = %[2]q
  schedule    = "cron(0 16 ? * TUE *)"
}
`, rName, desc)
}

func testAccMaintenanceWindowTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMaintenanceWindowTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccMaintenanceWindowCutoffConfig(rName string, cutoff int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = %d
  duration = 3
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, cutoff, rName)
}

func testAccMaintenanceWindowDurationConfig(rName string, duration int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = %d
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, duration, rName)
}

func testAccMaintenanceWindowEnabledConfig(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  enabled  = %t
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, enabled, rName)
}

func testAccMaintenanceWindowEndDateConfig(rName, endDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  end_date = %q
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, endDate, rName)
}

func testAccMaintenanceWindowMultipleUpdatesConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 8
  duration = 10
  enabled  = false
  name     = %q
  schedule = "cron(0 16 ? * WED *)"
}
`, rName)
}

func testAccMaintenanceWindowScheduleConfig(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %q
  schedule = %q
}
`, rName, schedule)
}

func testAccMaintenanceWindowScheduleTimezoneConfig(rName, scheduleTimezone string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff            = 1
  duration          = 3
  name              = %q
  schedule          = "cron(0 16 ? * TUE *)"
  schedule_timezone = %q
}
`, rName, scheduleTimezone)
}

func testAccMaintenanceWindowScheduleOffsetConfig(rName string, scheduleOffset int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff          = 1
  duration        = 3
  name            = %q
  schedule        = "cron(0 16 ? * TUE#3 *)"
  schedule_offset = %d
}
`, rName, scheduleOffset)
}

func testAccMaintenanceWindowStartDateConfig(rName, startDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff     = 1
  duration   = 3
  name       = %q
  schedule   = "cron(0 16 ? * TUE *)"
  start_date = %q
}
`, rName, startDate)
}
