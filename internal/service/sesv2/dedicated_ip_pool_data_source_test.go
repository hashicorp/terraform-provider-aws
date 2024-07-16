// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2DedicatedIPPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_sesv2_dedicated_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDedicatedIPPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "pool_name", rName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "ses", regexache.MustCompile(`dedicated-ip-pool/.+`)),
				),
			},
		},
	})
}

func testAccDedicatedIPPoolDataSourceConfig_basic(poolName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = %[1]q
}

data "aws_sesv2_dedicated_ip_pool" "test" {
  depends_on = [aws_sesv2_dedicated_ip_pool.test]
  pool_name  = %[1]q
}
`, poolName, poolName)
}
