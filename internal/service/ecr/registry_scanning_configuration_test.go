// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccECRScanningConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":  testAccRegistryScanningConfiguration_basic,
		"update": testAccRegistryScanningConfiguration_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccRegistryScanningConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.GetRegistryScanningConfigurationOutput
	resourceName := "aws_ecr_registry_scanning_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryScanningConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryScanningConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccRegistryScanningConfigurationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scan_type", "BASIC"),
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

func testAccRegistryScanningConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.GetRegistryScanningConfigurationOutput
	resourceName := "aws_ecr_registry_scanning_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryScanningConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryScanningConfigurationConfig_oneRule(),
				Check: resource.ComposeTestCheckFunc(
					testAccRegistryScanningConfigurationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"scan_frequency": "SCAN_ON_PUSH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						"filter":      "example",
						"filter_type": "WILDCARD",
					}),
					resource.TestCheckResourceAttr(resourceName, "scan_type", "BASIC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryScanningConfigurationConfig_twoRules(),
				Check: resource.ComposeTestCheckFunc(
					testAccRegistryScanningConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"scan_frequency": "CONTINUOUS_SCAN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						"filter":      "example",
						"filter_type": "WILDCARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"scan_frequency": "SCAN_ON_PUSH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						"filter":      "*",
						"filter_type": "WILDCARD",
					}),
					resource.TestCheckResourceAttr(resourceName, "scan_type", "ENHANCED"),
				),
			},
		},
	})
}

func testAccCheckRegistryScanningConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_registry_scanning_configuration" {
				continue
			}

			_, err := conn.GetRegistryScanningConfigurationWithContext(ctx, &ecr.GetRegistryScanningConfigurationInput{})

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccRegistryScanningConfigurationExists(ctx context.Context, name string, v *ecr.GetRegistryScanningConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Registry Scanning Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

		output, err := conn.GetRegistryScanningConfigurationWithContext(ctx, &ecr.GetRegistryScanningConfigurationInput{})

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRegistryScanningConfigurationConfig_basic() string {
	return `
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "BASIC"
}
`
}

func testAccRegistryScanningConfigurationConfig_oneRule() string {
	return `
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "BASIC"
  rule {
    scan_frequency = "SCAN_ON_PUSH"
    repository_filter {
      filter      = "example"
      filter_type = "WILDCARD"
    }
  }
}
`
}

func testAccRegistryScanningConfigurationConfig_twoRules() string {
	return `
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "ENHANCED"
  rule {
    scan_frequency = "CONTINUOUS_SCAN"
    repository_filter {
      filter      = "example"
      filter_type = "WILDCARD"
    }
  }
  rule {
    scan_frequency = "SCAN_ON_PUSH"
    repository_filter {
      filter      = "*"
      filter_type = "WILDCARD"
    }
  }
}
`
}
