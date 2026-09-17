// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package account_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAccountRegionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_account_regions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: `data "aws_account_regions" "test" {}`,
				Check:  resource.TestCheckResourceAttrSet(dataSourceName, "regions.#"),
			},
			{
				Config: testAccRegionsDataSourceConfig_status(string(awstypes.RegionOptStatusEnabledByDefault)),
				Check: resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
					"region_opt_status": string(awstypes.RegionOptStatusEnabledByDefault),
				}),
			},
			{
				Config: testAccRegionsDataSourceConfig_status(string(awstypes.RegionOptStatusDisabled)),
				Check: resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
					"region_opt_status": string(awstypes.RegionOptStatusDisabled),
				}),
			},
		},
	})
}

func testAccRegionsDataSourceConfig_status(status string) string {
	return fmt.Sprintf(`
data "aws_account_regions" "test" {
  region_opt_status_contains = ["%s"]
}
`, status)
}
