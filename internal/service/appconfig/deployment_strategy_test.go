// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigDeploymentStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentStrategyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`deploymentstrategy/[0-9a-z]{4,7}`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_duration_in_minutes", "3"),
					resource.TestCheckResourceAttr(resourceName, "growth_factor", "10"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "replicate_to", string(awstypes.ReplicateToNone)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAppConfigDeploymentStrategy_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentStrategyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				Config: testAccDeploymentStrategyConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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

func TestAccAppConfigDeploymentStrategy_updateFinalBakeTime(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentStrategyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_finalBakeTime(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_bake_time_in_minutes", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentStrategyConfig_finalBakeTime(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_bake_time_in_minutes", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test FinalBakeTimeInMinutes Removal
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAppConfigDeploymentStrategy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentStrategyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceDeploymentStrategy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeploymentStrategyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_deployment_strategy" {
				continue
			}

			_, err := tfappconfig.FindDeploymentStrategyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppConfig Deployment Strategy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentStrategyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		_, err := tfappconfig.FindDeploymentStrategyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDeploymentStrategyConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName)
}

func testAccDeploymentStrategyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  description                    = %[2]q
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName, description)
}

func testAccDeploymentStrategyConfig_finalBakeTime(rName string, time int) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = %[2]d
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName, time)
}
