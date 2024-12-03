// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMemoryDBMultiRegionCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMultiRegionClusterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Multi Region Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBClient(ctx)

		_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccCheckMultiRegionClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_memorydb_multi_region_cluster" {
				continue
			}

			_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MemoryDB Multi Region Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMultiRegionClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  suffix     = %[1]q
  node_type  = "db.r7g.xlarge"

  tags = {
    Test = "test"
  }
}
`, rName)
}
