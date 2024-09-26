// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2PublicIPv4PoolsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_public_ipv4_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testPublicIPv4PoolsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "pool_ids.#"),
				),
			},
		},
	})
}

func TestAccEC2PublicIPv4PoolsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_public_ipv4_pools.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testPublicIPv4PoolsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "pool_ids.#", acctest.Ct0),
				),
			},
		},
	})
}

const testPublicIPv4PoolsDataSourceConfig_basic = `
data "aws_ec2_public_ipv4_pools" "test" {}
`

func testPublicIPv4PoolsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_public_ipv4_pools" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
