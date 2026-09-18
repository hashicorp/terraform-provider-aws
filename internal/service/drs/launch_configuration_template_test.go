// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package drs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdrs "github.com/hashicorp/terraform-provider-aws/internal/service/drs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDRSLaunchConfigurationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_drs_launch_configuration_template.test"
	var lct awstypes.LaunchConfigurationTemplate

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DRSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationTemplateConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchConfigurationTemplateExists(ctx, t, resourceName, &lct),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "drs", "launch-configuration-template/{id}"),
					resource.TestCheckResourceAttr(resourceName, "copy_private_ip", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "copy_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "launch_disposition", "STARTED"),
					resource.TestCheckResourceAttr(resourceName, "launch_into_source_instance", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "licensing.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "licensing.0.os_byol", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "post_launch_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "recovery_mode", "OPTIMAL"),
					resource.TestCheckResourceAttr(resourceName, "target_instance_type_right_sizing_method", "IN_AWS"),
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

func TestAccDRSLaunchConfigurationTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_drs_launch_configuration_template.test"
	var lct awstypes.LaunchConfigurationTemplate

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DRSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationTemplateConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationTemplateExists(ctx, t, resourceName, &lct),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdrs.ResourceLaunchConfigurationTemplate, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDRSLaunchConfigurationTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_drs_launch_configuration_template.test"
	var lct awstypes.LaunchConfigurationTemplate

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DRSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationTemplateConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchConfigurationTemplateExists(ctx, t, resourceName, &lct),
					resource.TestCheckResourceAttr(resourceName, "copy_tags", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchConfigurationTemplateConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchConfigurationTemplateExists(ctx, t, resourceName, &lct),
					resource.TestCheckResourceAttr(resourceName, "copy_tags", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckLaunchConfigurationTemplateExists(ctx context.Context, t *testing.T, n string, v *awstypes.LaunchConfigurationTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DRSClient(ctx)

		output, err := tfdrs.FindLaunchConfigurationTemplateByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLaunchConfigurationTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DRSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_drs_launch_configuration_template" {
				continue
			}

			_, err := tfdrs.FindLaunchConfigurationTemplateByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DRS Launch Configuration Template (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLaunchConfigurationTemplateConfig_basic() string {
	return `
resource "aws_drs_launch_configuration_template" "test" {
  copy_private_ip             = true
  copy_tags                   = false
  launch_disposition          = "STARTED"
  launch_into_source_instance = false
  post_launch_enabled         = false
  recovery_mode               = "OPTIMAL"

  licensing {
    os_byol = true
  }

  target_instance_type_right_sizing_method = "IN_AWS"
}
`
}

func testAccLaunchConfigurationTemplateConfig_updated() string {
	return `
resource "aws_drs_launch_configuration_template" "test" {
  copy_private_ip             = true
  copy_tags                   = true
  launch_disposition          = "STARTED"
  launch_into_source_instance = false
  post_launch_enabled         = false
  recovery_mode               = "OPTIMAL"

  licensing {
    os_byol = true
  }

  target_instance_type_right_sizing_method = "IN_AWS"
}
`
}
