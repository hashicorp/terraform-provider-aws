// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPWorkspaceConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeWorkspaceConfigurationOutput
	resourceName := "aws_prometheus_workspace_configuration.test"
	workspaceResourceName := "aws_prometheus_workspace.test"
	retentionPeriod := 30
	retentionPeriodUpdated := 15

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigurationConfig_basic(retentionPeriod),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", strconv.Itoa(retentionPeriod)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"limits_per_label_set", "retention_period_in_days"},
			},
			{
				Config: testAccWorkspaceConfigurationConfig_basic(retentionPeriodUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", strconv.Itoa(retentionPeriodUpdated)),
				),
			},
		},
	})
}

func TestAccAMPWorkspaceConfiguration_defaultBucket(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeWorkspaceConfigurationOutput
	resourceName := "aws_prometheus_workspace_configuration.test"
	workspaceResourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigurationConfig_defaultBucket(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"limits_per_label_set", "retention_period_in_days"},
			},
		},
	})
}
func TestAccAMPWorkspaceConfiguration_limitPerLabelSet(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeWorkspaceConfigurationOutput
	resourceName := "aws_prometheus_workspace_configuration.test"
	workspaceResourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigurationConfig_limitPerLabelSet(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.0.label_set.__name__", "rest_client_request_duration_seconds_bucket"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.1.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.1.label_set.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.1.limits.0.max_series", "10000"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.2.limits.0.max_series", "60000"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", "30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"limits_per_label_set", "retention_period_in_days"},
			},
		},
	})
}

func testAccCheckWorkspaceConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_workspace_configuration" {
				continue
			}

			_, err := tfamp.FindWorkspaceConfigurationByID(ctx, conn, rs.Primary.ID)

			if err == nil {
				return fmt.Errorf("AMP Workspace Configuration %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckWorkspaceConfigurationExists(ctx context.Context, n string, v *amp.DescribeWorkspaceConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		output, err := tfamp.FindWorkspaceConfigurationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWorkspaceConfigurationConfig_basic(retentionPeriod int) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = %[1]d
}
`, retentionPeriod)
}

func testAccWorkspaceConfigurationConfig_defaultBucket() string {
	return `
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id

  limits_per_label_set {
    label_set = {}
	limits {
	  max_series = 100000
	}
  }
}
`
}

func testAccWorkspaceConfigurationConfig_limitPerLabelSet() string {
	return `
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = 30

  limits_per_label_set {
    label_set = {
		"__name__" = "rest_client_request_duration_seconds_bucket"
		"region" = "us-east-1"
	}
	limits {
	  max_series = 1000
	}
  }

  limits_per_label_set {
    label_set = {
		"env" = "dev"
	}
	limits {
	  max_series = 10000
	}
  }

  limits_per_label_set {
    label_set = {
		"env" = "prod"
	}
	limits {
	  max_series = 60000
	}
  }
}
`
}
