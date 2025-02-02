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

func TestAccDBProxiesDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_proxies.test"
	resourceName := "aws_db_proxy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxiesDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "proxy_arns.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "proxy_arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "proxy_names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccProxiesDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_basic(rName), testAccProxyConfig_basic(rName), `
data "aws_db_Proxies "test" {}
`)
}
