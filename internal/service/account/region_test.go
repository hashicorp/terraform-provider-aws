// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRegionDisabled(ctx, t, "", regionName) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccRegionConfig_basic(regionName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
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
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "opt_status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "region_name", regionName),
				),
			},
		},
	})
}

func testAccPreCheckRegionDisabled(ctx context.Context, t *testing.T, accountID, region string) {
	t.Helper()

	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

	output, err := tfaccount.FindRegionOptStatus(ctx, conn, accountID, region)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if status := output.RegionOptStatus; status != types.RegionOptStatusDisabled {
		t.Skipf("unexpected status (%s): %s", region, status)
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
