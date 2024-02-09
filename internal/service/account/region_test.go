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
	region := "ap-southeast-3" //lintignore:AWSAT003

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(region, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrimaryContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
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

		_, err := tfaccount.FindRegionOptInStatus(ctx, conn, rs.Primary.Attributes["account_id"], rs.Primary.Attributes["region"])

		return err
	}
}

func testAccRegionConfig_basic(region string, enabled string) string {
	return fmt.Sprintf(`
resource "aws_account_region" "test" {
  region  = %[1]q
  enabled = %[2]q
}
`, region, enabled)
}
