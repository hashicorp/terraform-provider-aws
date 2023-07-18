// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcontroltower "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccControlTowerControl_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Control": {
			"basic":      testAccControl_basic,
			"disappears": testAccControl_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var control controltower.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	controlName := "AWS-GR_EC2_VOLUME_INUSE_CHECK"
	ouName := "Security"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		CheckDestroy:             testAccCheckControlDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(controlName, ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
				),
			},
		},
	})
}

func testAccControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var control controltower.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	controlName := "AWS-GR_EC2_VOLUME_INUSE_CHECK"
	ouName := "Security"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(controlName, ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, resourceName, &control),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcontroltower.ResourceControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckControlExists(ctx context.Context, n string, v *controltower.EnabledControlSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ControlTower Control ID is set")
		}

		targetIdentifier, controlIdentifier, err := tfcontroltower.ControlParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerConn(ctx)

		output, err := tfcontroltower.FindEnabledControlByTwoPartKey(ctx, conn, targetIdentifier, controlIdentifier)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckControlDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_control" {
				continue
			}

			targetIdentifier, controlIdentifier, err := tfcontroltower.ControlParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfcontroltower.FindEnabledControlByTwoPartKey(ctx, conn, targetIdentifier, controlIdentifier)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ControlTower Control %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccControlConfig_basic(controlName string, ouName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_controltower_control" "test" {
  control_identifier = "arn:${data.aws_partition.current.partition}:controltower:${data.aws_region.current.name}::control/%[1]s"
  target_identifier = [
    for x in data.aws_organizations_organizational_units.test.children :
    x.arn if x.name == "%[2]s"
  ][0]
}
`, controlName, ouName)
}
