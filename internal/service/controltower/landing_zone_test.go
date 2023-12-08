// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcontroltower "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLandingZoneTowerLandingZone_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"LandingZone": {
			"basic":      testAccLandingZone_basic,
			"disappears": testAccLandingZone_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccLandingZone_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_controltower_landing_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerEndpointID),
		CheckDestroy:             testAccCheckLandingZoneDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLandingZoneConfig_basic("1.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0"),
				),
			},
		},
	})
}

func testAccLandingZone_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_controltower_landing_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLandingZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLandingZoneConfig_basic("1.0"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcontroltower.ResourceControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLandingZoneDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_landing_zone" {
				continue
			}

			input := &controltower.GetLandingZoneInput{
				LandingZoneIdentifier: &rs.Primary.ID,
			}

			_, err := conn.GetLandingZone(ctx, input)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Control Tower Landing Zone %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLandingZoneConfig_basic(version string) string {
	return fmt.Sprintf(`
resource "aws_controltower_landing_zone" "test" {
  manifest = jsondecode(file("${path.module}/fixtures/LandingZoneManifest.json"))
  version  = "%[1]s"
}
`, version)
}
