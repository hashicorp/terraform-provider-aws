// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPAlertManagerDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_prometheus_alert_manager_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertManagerDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", defaultAlertManagerDefinition()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAlertManagerDefinitionConfig_basic(anotherAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", anotherAlertManagerDefinition()),
				),
			},
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "definition", defaultAlertManagerDefinition()),
				),
			},
		},
	})
}

func TestAccAMPAlertManagerDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_prometheus_alert_manager_definition.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertManagerDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlertManagerDefinitionConfig_basic(defaultAlertManagerDefinition()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertManagerDefinitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamp.ResourceAlertManagerDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAlertManagerDefinitionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		_, err := tfamp.FindAlertManagerDefinitionByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckAlertManagerDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_alert_manager_definition" {
				continue
			}

			_, err := tfamp.FindAlertManagerDefinitionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Alert Manager Definition %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func defaultAlertManagerDefinition() string {
	return `
alertmanager_config: |
  route:
    receiver: 'default'
  receivers:
    - name: 'default'
`
}

func anotherAlertManagerDefinition() string {
	return `
alertmanager_config: |
  route:
    receiver: 'default2'
  receivers:
    - name: 'default2'
`
}

func testAccAlertManagerDefinitionConfig_basic(definition string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {}

resource "aws_prometheus_alert_manager_definition" "test" {
  workspace_id = aws_prometheus_workspace.test.id
  definition   = %[1]q
}
`, definition)
}
