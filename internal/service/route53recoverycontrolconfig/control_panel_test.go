// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoverycontrolconfig "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccControlPanel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlPanelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccControlPanelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
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

func testAccControlPanel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlPanelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccControlPanelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53recoverycontrolconfig.ResourceControlPanel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckControlPanelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53recoverycontrolconfig_control_panel" {
				continue
			}

			_, err := tfroute53recoverycontrolconfig.FindControlPanelByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53RecoveryControlConfig Control Panel (%s) not deleted", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckControlPanelExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

		_, err := tfroute53recoverycontrolconfig.FindControlPanelByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccClusterSetUp(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccControlPanelConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterSetUp(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName))
}
