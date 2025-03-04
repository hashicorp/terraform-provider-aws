// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAccountRegionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_account_regions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountRegionsDataSourceConfig_inAccount(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
						"region_name":       "us-east-1",
						"region_opt_status": "ENABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
						"region_name":       "us-east-2",
						"region_opt_status": "ENABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
						"region_name":       "us-west-1",
						"region_opt_status": "ENABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
						"region_name":       "us-west-2",
						"region_opt_status": "ENABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
						"region_name":       "af-south-1",
						"region_opt_status": "DISABLED",
					}),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", "data.aws_caller_identity.current", "account_id"),
				),
			},
		},
	})
}

func testAccAccountRegionsDataSourceConfig_inAccount() string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_account_regions" "test" {}
`)
}
