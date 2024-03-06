// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_region.test"
	regionName := names.APSoutheast3RegionID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(regionName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRegionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "opt_status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "region_name", regionName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegionConfig_basic(regionName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRegionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "opt_status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "region_name", regionName),
				),
			},
		},
	})
}

func testAccRegionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

		_, err := tfaccount.FindRegionOptStatus(ctx, conn, rs.Primary.Attributes["account_id"], rs.Primary.Attributes["region_name"])

		return err
	}
}

func testAccRegionConfig_basic(region string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_account_region" "test" {
  region_name = %[1]q
  enabled     = %[2]t
}
`, region, enabled)
}
