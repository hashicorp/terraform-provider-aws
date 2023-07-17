// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "vpc-lattice", regexp.MustCompile(`servicenetwork/sn-.+`)),
				),
			},
		},
	})
}

func TestAccVPCLatticeServiceNetworkDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagKey := "tag1"
	tagValue := "value1"
	dataSourceName := "data.aws_vpclattice_service_network.test_tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkDataSourceConfig_tags(rName, tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "value1"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "vpc-lattice", regexp.MustCompile(`servicenetwork/sn-.+`)),
				),
			},
		},
	})
}

func testAccServiceNetworkDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`  
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

data "aws_vpclattice_service_network" "test" {
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName)
}

func testAccServiceNetworkDataSourceConfig_tags(rName string, tagKey string, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test_tags" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_vpclattice_service_network" "test_tags" {
  service_network_identifier = aws_vpclattice_service_network.test_tags.id
}
`, rName, tagKey, tagValue)
}
