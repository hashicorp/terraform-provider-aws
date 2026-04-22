// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53recoverycontrolconfig "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoutingControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoutingControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_arn", // not available in DescribeRoutingControlOutput
				},
			},
		},
	})
}

func testAccRoutingControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoutingControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53recoverycontrolconfig.ResourceRoutingControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoutingControl_nonDefaultControlPanel(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoutingControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingControlConfig_inNonDefaultPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingControlExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
				),
			},
		},
	})
}

func testAccCheckRoutingControlExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53RecoveryControlConfigClient(ctx)

		_, err := tfroute53recoverycontrolconfig.FindRoutingControlByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckRoutingControlDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53RecoveryControlConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53recoverycontrolconfig_routing_control" {
				continue
			}

			_, err := tfroute53recoverycontrolconfig.FindRoutingControlByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53RecoveryControlConfig Routing Control (%s) not deleted", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRoutingControlConfig_inDefaultPanel(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBase(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName))
}

func testAccControlPanelBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName)
}

func testAccRoutingControlConfig_inNonDefaultPanel(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBase(rName),
		testAccControlPanelBase(rName),
		fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[1]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
}
`, rName))
}
