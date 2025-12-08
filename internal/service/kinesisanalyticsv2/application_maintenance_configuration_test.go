// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesisanalyticsv2 "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesisanalyticsv2_application_maintenance_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationMaintenanceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "application_maintenance_window_start_time", "02:00"),
					resource.TestCheckResourceAttrSet(resourceName, "original_maintenance_window_start_time"),
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

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesisanalyticsv2_application_maintenance_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationMaintenanceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_maintenance_window_start_time", "02:00"),
				),
			},
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "03:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_maintenance_window_start_time", "03:30"),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesisanalyticsv2_application_maintenance_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationMaintenanceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesisanalyticsv2.ResourceApplicationMaintenanceConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_Disappears_application(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesisanalyticsv2_application_maintenance_configuration.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationMaintenanceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesisanalyticsv2.ResourceApplication(), applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationMaintenanceConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesisanalyticsv2_application_maintenance_configuration" {
				continue
			}

			application, err := tfkinesisanalyticsv2.FindApplicationDetailByName(ctx, conn, rs.Primary.ID)
			if err != nil {
				return err
			}

			// Check if maintenance window was restored to original
			if application.ApplicationMaintenanceConfigurationDescription != nil {
				currentTime := application.ApplicationMaintenanceConfigurationDescription.ApplicationMaintenanceWindowStartTime
				originalTime := rs.Primary.Attributes["original_maintenance_window_start_time"]
				if *currentTime != originalTime {
					return fmt.Errorf("Kinesis Analytics v2 Application Maintenance Configuration not restored to original")
				}
			}
		}
		return nil
	}
}

func testAccCheckApplicationMaintenanceConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

		_, err := tfkinesisanalyticsv2.FindApplicationDetailByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccApplicationMaintenanceConfigurationConfig_basic(rName, maintenanceTime string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_applicationMode(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application_maintenance_configuration" "test" {
  application_name                          = aws_kinesisanalyticsv2_application.test.name
  application_maintenance_window_start_time = %[1]q
}
`, maintenanceTime))
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_invalidTimeFormat(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(context.Background(), t); testAccPreCheck(context.Background(), t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationMaintenanceConfigurationConfig_basic(rName, "25:00"),
				ExpectError: regexache.MustCompile(`must be in HH:MM format \(00:00-23:59\)`),
			},
			{
				Config:      testAccApplicationMaintenanceConfigurationConfig_basic(rName, "12:60"),
				ExpectError: regexache.MustCompile(`must be in HH:MM format \(00:00-23:59\)`),
			},
			{
				Config:      testAccApplicationMaintenanceConfigurationConfig_basic(rName, "1200"),
				ExpectError: regexache.MustCompile(`must be in HH:MM format \(00:00-23:59\)`),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_invalidApplicationName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(context.Background(), t); testAccPreCheck(context.Background(), t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationMaintenanceConfigurationConfig_invalidName("invalid@name", "12:00"),
				ExpectError: regexache.MustCompile(`must only include alphanumeric, underscore, period, or hyphen characters`),
			},
		},
	})
}

func testAccApplicationMaintenanceConfigurationConfig_invalidName(appName, maintenanceTime string) string {
	return fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application_maintenance_configuration" "test" {
  application_name                          = %[1]q
  application_maintenance_window_start_time = %[2]q
}
`, appName, maintenanceTime)
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_runningApplication(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesisanalyticsv2_application_maintenance_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationMaintenanceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_runningApplication(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "application_maintenance_window_start_time", "02:00"),
				),
			},
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_runningApplication(rName, "03:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_maintenance_window_start_time", "03:30"),
				),
			},
		},
	})
}

func testAccApplicationMaintenanceConfigurationConfig_runningApplication(rName, maintenanceTime string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_flinkConfiguration(rName, acctest.CtTrue, "FLINK-1_20"),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application_maintenance_configuration" "test" {
  application_name                          = aws_kinesisanalyticsv2_application.test.name
  application_maintenance_window_start_time = %[1]q
}
`, maintenanceTime))
}

func TestAccKinesisAnalyticsV2ApplicationMaintenanceConfiguration_applicationNotReady(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationMaintenanceConfigurationConfig_basic(rName, "02:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationMaintenanceConfigurationExists(ctx, "aws_kinesisanalyticsv2_application_maintenance_configuration.test"),
					testAccCheckApplicationStop(ctx, rName),
				),
			},
			{
				Config:      testAccApplicationMaintenanceConfigurationConfig_basic(rName, "03:30"),
				ExpectError: regexache.MustCompile(`application must be in READY or RUNNING state`),
			},
		},
	})
}

func testAccCheckApplicationStop(ctx context.Context, applicationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

		_, err := conn.StopApplication(ctx, &kinesisanalyticsv2.StopApplicationInput{
			ApplicationName: aws.String(applicationName),
		})

		return err
	}
}
