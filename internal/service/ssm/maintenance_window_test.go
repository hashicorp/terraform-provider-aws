// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSSMMaintenanceWindow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var winId ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
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
	ctx := acctest.Context(t)
	var winId ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_description(rName, "foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
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
				Config: testAccMaintenanceWindowConfig_description(rName, "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "bar"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var winId ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
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
				Config: testAccMaintenanceWindowConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var winId ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &winId),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceMaintenanceWindow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_multipleUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig_multipleUpdates(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
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
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_cutoff(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_cutoff(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "2"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_duration(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_duration(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_duration(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_endDate(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.GetMaintenanceWindowOutput
	endDate1 := time.Now().UTC().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	endDate2 := time.Now().UTC().Add(730 * 24 * time.Hour).Format(time.RFC3339)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_endDate(rName, endDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_endDate(rName, endDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate2),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_schedule(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_schedule(rName, "cron(0 16 ? * TUE *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_schedule(rName, "cron(0 16 ? * WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_scheduleTimezone(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_scheduleTimezone(rName, "America/Los_Angeles"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/Los_Angeles"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_scheduleTimezone(rName, "America/New_York"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/New_York"),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", ""),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_scheduleOffset(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2 ssm.GetMaintenanceWindowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_scheduleOffset(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule_offset", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_scheduleOffset(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule_offset", "5"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindow_startDate(t *testing.T) {
	ctx := acctest.Context(t)
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.GetMaintenanceWindowOutput
	startDate1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	startDate2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowConfig_startDate(rName, startDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowConfig_startDate(rName, startDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate2),
				),
			},
			{
				Config: testAccMaintenanceWindowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowExists(ctx, resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
				),
			},
		},
	})
}

func testAccCheckMaintenanceWindowExists(ctx context.Context, n string, v *ssm.GetMaintenanceWindowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		output, err := tfssm.FindMaintenanceWindowByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMaintenanceWindowDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_maintenance_window" {
				continue
			}

			_, err := tfssm.FindMaintenanceWindowByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Maintenance Window %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMaintenanceWindowConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName)
}

func testAccMaintenanceWindowConfig_description(rName, desc string) string {
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

func testAccMaintenanceWindowConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccMaintenanceWindowConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccMaintenanceWindowConfig_cutoff(rName string, cutoff int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = %[2]d
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName, cutoff)
}

func testAccMaintenanceWindowConfig_duration(rName string, duration int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = %[2]d
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName, duration)
}

func testAccMaintenanceWindowConfig_enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  enabled  = %[2]t
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName, enabled)
}

func testAccMaintenanceWindowConfig_endDate(rName, endDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  end_date = %[2]q
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName, endDate)
}

func testAccMaintenanceWindowConfig_multipleUpdates(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 8
  duration = 10
  enabled  = false
  name     = %[1]q
  schedule = "cron(0 16 ? * WED *)"
}
`, rName)
}

func testAccMaintenanceWindowConfig_schedule(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = %[2]q
}
`, rName, schedule)
}

func testAccMaintenanceWindowConfig_scheduleTimezone(rName, scheduleTimezone string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff            = 1
  duration          = 3
  name              = %[1]q
  schedule          = "cron(0 16 ? * TUE *)"
  schedule_timezone = %[2]q
}
`, rName, scheduleTimezone)
}

func testAccMaintenanceWindowConfig_scheduleOffset(rName string, scheduleOffset int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff          = 1
  duration        = 3
  name            = %[1]q
  schedule        = "cron(0 16 ? * TUE#3 *)"
  schedule_offset = %[2]d
}
`, rName, scheduleOffset)
}

func testAccMaintenanceWindowConfig_startDate(rName, startDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff     = 1
  duration   = 3
  name       = %[1]q
  schedule   = "cron(0 16 ? * TUE *)"
  start_date = %[2]q
}
`, rName, startDate)
}
