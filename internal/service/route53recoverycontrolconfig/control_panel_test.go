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

func testAccControlPanel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlPanelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlPanelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
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

func testAccControlPanel_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlPanelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlPanelConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccControlPanelConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccControlPanelConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}
func testAccControlPanel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryControlConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlPanelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlPanelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlPanelExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53recoverycontrolconfig.ResourceControlPanel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckControlPanelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53RecoveryControlConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53recoverycontrolconfig_control_panel" {
				continue
			}

			_, err := tfroute53recoverycontrolconfig.FindControlPanelByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckControlPanelExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53RecoveryControlConfigClient(ctx)

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

func testAccControlPanelConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterSetUp(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccControlPanelConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterSetUp(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
