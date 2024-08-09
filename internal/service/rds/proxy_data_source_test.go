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

func TestAccRDSProxyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_db_proxy.test"
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "auth.#", resourceName, "auth.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "debug_logging", resourceName, "debug_logging"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_family", resourceName, "engine_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_client_timeout", resourceName, "idle_client_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "require_tls", resourceName, "require_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrRoleARN, resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_security_group_ids.#", resourceName, "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_subnet_ids.#", resourceName, "vpc_subnet_ids.#"),
				),
			},
		},
	})
}

func testAccProxyDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_basic(rName), `
data "aws_db_proxy" "test" {
  name = aws_db_proxy.test.name
}
`)
}
