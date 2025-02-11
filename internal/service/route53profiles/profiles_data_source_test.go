// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesProfilesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_profiles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfilesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfilesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "profiles.#", 0),
				),
			},
		},
	})
}

func testAccCheckProfilesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_profile" {
				continue
			}

			_, err := tfroute53profiles.FindProfileByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("route 53 Profile %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccProfilesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

data "aws_route53profiles_profiles" "test" {
  depends_on = [aws_route53profiles_profile.test]
}
`, rName)
}
