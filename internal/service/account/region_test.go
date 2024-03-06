// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAccountRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_region.test"
	regionName := "ap-southeast-3" //lintignore:AWSAT003

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(regionName, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRegionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "opt_status", regexache.MustCompile(`ENABLED|ENABLING`)),
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

func TestAccAccountRegion_enabledByDefault(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_region.test"
	regionName := "us-east-1" //lintignore:AWSAT003

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(regionName, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRegionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "opt_status", regexache.MustCompile(`ENABLED_BY_DEFAULT`)),
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

func testAccRegionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account Region ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

		_, err := tfaccount.FindRegionOptStatus(ctx, conn, rs.Primary.Attributes["account_id"], rs.Primary.Attributes["region_name"])

		return err
	}
}

func testAccRegionConfig_basic(region string, enabled string) string {
	return fmt.Sprintf(`
resource "aws_account_region" "test" {
  region_name = %[1]q
  enabled     = %[2]q
}
`, region, enabled)
}
