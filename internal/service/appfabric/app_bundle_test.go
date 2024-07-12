// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppBundle_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					resource.TestCheckNoResourceAttr(resourceName, "customer_managed_key_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccAppBundle_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppBundle, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAppBundle_cmk(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_cmk(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					resource.TestCheckResourceAttrSet(resourceName, "customer_managed_key_arn"),
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

func testAccAppBundle_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
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
				Config: testAccAppBundleConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppBundleConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAppBundleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_bundle" {
				continue
			}

			_, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppFabric App Bundle %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppBundleExists(ctx context.Context, n string, v *awstypes.AppBundle) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		output, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccAppBundleConfig_basic = `
resource "aws_appfabric_app_bundle" "test" {}
`

func testAccAppBundleConfig_cmk(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_appfabric_app_bundle" "test" {
  customer_managed_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccAppBundleConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAppBundleConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
