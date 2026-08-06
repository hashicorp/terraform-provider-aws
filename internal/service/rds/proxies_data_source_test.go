// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDBProxiesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_proxies.test"
	resourceName := "aws_db_proxy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxiesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccProxiesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_basic(rName), `
data "aws_db_proxies" "test" {
  depends_on = [aws_db_proxy.test]
}`)
}
