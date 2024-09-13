// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2PublicIPv4PoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_public_ipv4_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPublicIPv4Pools(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testPublicIPv4PoolDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_address_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_available_address_count"),
				),
			},
		},
	})
}

func testAccPreCheckPublicIPv4Pools(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	output, err := tfec2.FindPublicIPv4Pools(ctx, conn, &ec2.DescribePublicIpv4PoolsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is at least one pool.
	if len(output) == 0 {
		t.Skip("skipping since no EC2 Public IPv4 Pools found")
	}
}

const testPublicIPv4PoolDataSourceConfig_basic = `
data "aws_ec2_public_ipv4_pools" "test" {}

data "aws_ec2_public_ipv4_pool" "test" {
  pool_id = data.aws_ec2_public_ipv4_pools.test.pool_ids[0]
}
`
