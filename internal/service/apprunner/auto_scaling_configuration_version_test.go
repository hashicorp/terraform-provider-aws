// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerAutoScalingConfigurationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "has_associated_service", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
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

func TestAccAppRunnerAutoScalingConfigurationVersion_complex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName, 50, 10, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "50"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName, 150, 20, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "150"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_multipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, otherResourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					acctest.MatchResourceAttrRegionalARN(ctx, otherResourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, names.AttrStatus, "active"),
				),
			},
			{
				// Test update of "latest" computed attribute after apply
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, otherResourceName),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtFalse),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_updateMultipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, otherResourceName),
				),
			},
			{
				Config: testAccAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, otherResourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					acctest.MatchResourceAttrRegionalARN(ctx, otherResourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "125"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapprunner.ResourceAutoScalingConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAutoScalingConfigurationVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_auto_scaling_configuration_version" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

			_, err := tfapprunner.FindAutoScalingConfigurationByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner AutoScaling Configuration Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAutoScalingConfigurationVersionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

		_, err := tfapprunner.FindAutoScalingConfigurationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAutoScalingConfigurationVersionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}
`, rName)
}

func testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName string, maxConcurrency, maxSize, minSize int) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q

  max_concurrency = %[2]d
  max_size        = %[3]d
  min_size        = %[4]d
}
`, rName, maxConcurrency, maxSize, minSize)
}

func testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name
}
`, rName)
}

func testAccAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name

  max_concurrency = 125
  max_size        = 20
}
`, rName)
}
