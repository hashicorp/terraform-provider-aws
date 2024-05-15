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

func testAccLandingZone_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_controltower_landing_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckNoLandingZone(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		CheckDestroy:             testAccCheckLandingZoneDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLandingZoneConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLandingZoneExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "drift_status.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "latest_available_version"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0"),
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

func testAccLandingZone_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_controltower_landing_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckNoLandingZone(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLandingZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLandingZoneConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLandingZoneExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcontroltower.ResourceLandingZone(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLandingZone_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_controltower_landing_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckNoLandingZone(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		CheckDestroy:             testAccCheckLandingZoneDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLandingZoneConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLandingZoneExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLandingZoneConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLandingZoneExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLandingZoneConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLandingZoneExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccPreCheckNoLandingZone(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

	input := &controltower.ListLandingZonesInput{}
	var n int
	pages := controltower.NewListLandingZonesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		n += len(page.LandingZones)
	}

	if n > 0 {
		t.Skip("skipping since Landing Zone already exists")
	}
}

func testAccCheckLandingZoneExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

		_, err := tfcontroltower.FindLandingZoneByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckLandingZoneDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_landing_zone" {
				continue
			}

			_, err := tfcontroltower.FindLandingZoneByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ControlTower Landing Zone %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const landingZoneVersion = "3.3"

var testAccLandingZoneConfig_basic = fmt.Sprintf(`
resource "aws_controltower_landing_zone" "test" {
  manifest_json = file("${path.module}/test-fixtures/LandingZoneManifest.json")

  version = %[2]q
}
`, acctest.Region(), landingZoneVersion)

func testAccLandingZoneConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_controltower_landing_zone" "test" {
  manifest_json = jfile("${path.module}/test-fixtures/LandingZoneManifest.json")

  version = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, acctest.Region(), landingZoneVersion, tagKey1, tagValue1)
}

func testAccLandingZoneConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_controltower_landing_zone" "test" {
  manifest_json = file("${path.module}/test-fixtures/LandingZoneManifest.json")

  version = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, acctest.Region(), landingZoneVersion, tagKey1, tagValue1, tagKey2, tagValue2)
}
