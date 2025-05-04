// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigurationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfigurationConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", "15"),
				),
			},
		},
	})
}

func TestAccAMPWorkspaceConfiguration_configuration(t *testing.T) {
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
				Config: testAccWorkspaceConfigurationConfig_configuration(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration"),
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

func TestAccAMPWorkspaceConfiguration_limitsPerLabelSet(t *testing.T) {
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
				Config: testAccWorkspaceConfigurationConfig_limitsPerLabelSet(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.0.label_set.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.0.label_set.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "limits_per_label_set.0.label_set.env", "dev"),
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

func testAccWorkspaceConfigurationConfig_basic() string {
	return `
resource "aws_prometheus_workspace" "test" {
  alias = "test"
}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = 8
}
`
}

func testAccWorkspaceConfigurationConfig_updated() string {
	return `
resource "aws_prometheus_workspace" "test" {
  alias = "test"
}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = 15
}
`
}

func testAccWorkspaceConfigurationConfig_configuration() string {
	return `
resource "aws_prometheus_workspace" "test" {
  alias = "test"
}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = 8
  configuration            = <<EOF
alertmanager_config: |
  route:
    receiver: 'default'
  receivers:
    - name: 'default'
EOF
}
`
}

func testAccWorkspaceConfigurationConfig_limitsPerLabelSet() string {
	return `
resource "aws_prometheus_workspace" "test" {
  alias = "test"
}

resource "aws_prometheus_workspace_configuration" "test" {
  workspace_id             = aws_prometheus_workspace.test.id
  retention_period_in_days = 8

  limits_per_label_set {
    label_set = {
      name = "test"
      env  = "dev"
    }
  }
}
`
}
