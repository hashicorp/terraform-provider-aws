// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_ec2ScanMode("EC2_HYBRID"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ec2_configuration.0.scan_mode", "EC2_HYBRID"),
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

func testAccConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Use a non-default scan mode so that Delete (which resets to
				// EC2_HYBRID, the AWS default) produces a state diff from the
				// declared config. Without this, refresh-after-disappear would
				// show no diff and the test would fail with an empty plan.
				Config: testAccConfigurationConfig_ec2ScanMode("EC2_SSM_AGENT_BASED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfinspector2.ResourceConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfiguration_ec2ScanMode(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_ec2ScanMode("EC2_SSM_AGENT_BASED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ec2_configuration.0.scan_mode", "EC2_SSM_AGENT_BASED"),
				),
			},
			{
				Config: testAccConfigurationConfig_ec2ScanMode("EC2_HYBRID"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ec2_configuration.0.scan_mode", "EC2_HYBRID"),
				),
			},
		},
	})
}

func testAccConfiguration_ecrRescan(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_ecrRescan("DAYS_14", "DAYS_14", "LAST_IN_USE_AT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.rescan_duration", "DAYS_14"),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.pull_date_rescan_duration", "DAYS_14"),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.pull_date_rescan_mode", "LAST_IN_USE_AT"),
				),
			},
			{
				Config: testAccConfigurationConfig_ecrRescan("DAYS_30", "DAYS_30", "LAST_PULL_DATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.rescan_duration", "DAYS_30"),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.pull_date_rescan_duration", "DAYS_30"),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.pull_date_rescan_mode", "LAST_PULL_DATE"),
				),
			},
		},
	})
}

func testAccConfiguration_combined(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_combined("EC2_HYBRID", "DAYS_14"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ec2_configuration.0.scan_mode", "EC2_HYBRID"),
					resource.TestCheckResourceAttr(resourceName, "ecr_configuration.0.rescan_duration", "DAYS_14"),
				),
			},
		},
	})
}

func testAccCheckConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_configuration" {
				continue
			}

			// The configuration resource always exists at the API level — Delete
			// resets to AWS defaults rather than removing. Verify the reset took.
			out, err := tfinspector2.FindConfiguration(ctx, conn)
			if err != nil {
				return err
			}

			if out.Ec2Configuration != nil && out.Ec2Configuration.ScanModeState != nil {
				if mode := out.Ec2Configuration.ScanModeState.ScanMode; mode != "EC2_HYBRID" {
					return fmt.Errorf("Inspector2 Configuration EC2 scan_mode = %q after delete; expected EC2_HYBRID (default)", mode)
				}
			}
		}

		return nil
	}
}

func testAccCheckConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		_, err := tfinspector2.FindConfiguration(ctx, conn)

		return err
	}
}

func testAccConfigurationConfig_ec2ScanMode(scanMode string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_configuration" "test" {
  ec2_configuration {
    scan_mode = %[1]q
  }
}
`, scanMode)
}

func testAccConfigurationConfig_ecrRescan(rescanDuration, pullDateDuration, pullDateMode string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_configuration" "test" {
  ecr_configuration {
    rescan_duration           = %[1]q
    pull_date_rescan_duration = %[2]q
    pull_date_rescan_mode     = %[3]q
  }
}
`, rescanDuration, pullDateDuration, pullDateMode)
}

func testAccConfigurationConfig_combined(scanMode, rescanDuration string) string {
	return fmt.Sprintf(`
resource "aws_inspector2_configuration" "test" {
  ec2_configuration {
    scan_mode = %[1]q
  }
  ecr_configuration {
    rescan_duration = %[2]q
  }
}
`, scanMode, rescanDuration)
}
