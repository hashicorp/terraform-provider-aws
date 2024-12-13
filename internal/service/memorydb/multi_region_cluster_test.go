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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_cluster_name"),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_cluster_name_suffix"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_parameter_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, "num_shards"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_defaults(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_cluster_name"),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_cluster_name_suffix"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_parameter_group_name"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_description(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_description(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Also managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
			{
				Config: testAccMultiRegionClusterConfig_description(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Also managed by Terraform, but now with an updated description"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_tlsEnabled(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_tlsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_engine(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_engine(rName, "valkey"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_engineVersion(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_engineVersion(rName, "valkey", "7.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "7.3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_updateStrategy(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_updateStrategy(rName, "coordinated", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "update_strategy", "coordinated"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_updateStrategy(rName, "uncoordinated", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "update_strategy", "uncoordinated"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll, "update_strategy"},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_numShards(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_numShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
			{
				Config: testAccMultiRegionClusterConfig_numShards(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "3"),
				),
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTags, names.AttrTagsAll},
			},
			{
				Config: testAccMultiRegionClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
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

		_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, rs.Primary.Attributes["multi_region_cluster_name"])

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

			_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, rs.Primary.Attributes["multi_region_cluster_name"])

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
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"

  tags = {
    Test = "test"
  }
}
`, rName)
}

func testAccMultiRegionClusterConfig_defaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
}
`, rName)
}

// Sets `num_shards` to also test an update of the resource
func testAccMultiRegionClusterConfig_description(rName string, numShards int) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  description = "Also managed by Terraform"

  num_shards = %[2]d

  tags = {
    Test = "test"
  }
}
`, rName, numShards)
}

func testAccMultiRegionClusterConfig_numShards(rName string, numShards int) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  description = "Also managed by Terraform"

  num_shards = %[2]d

  tags = {
    Test = "test"
  }
}
`, rName, numShards)
}

func testAccMultiRegionClusterConfig_tlsEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  tls_enabled = false

  tags = {
    Test = "test"
  }
}
`, rName)
}

func testAccMultiRegionClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  engine      = %[2]q

  tags = {
    Test = "test"
  }
}
`, rName, engine)
}

func testAccMultiRegionClusterConfig_engineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix    = %[1]q
  node_type      = "db.r7g.xlarge"
  engine         = %[2]q
  engine_version = %[3]q

  tags = {
    Test = "test"
  }
}
`, rName, engine, engineVersion)
}

// Sets `num_shards` to update the resource
func testAccMultiRegionClusterConfig_updateStrategy(rName, updateStrategy string, numShards int) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  
  update_strategy = %[2]q
  num_shards = %[3]d

  tags = {
    Test = "test"
  }
}
`, rName, updateStrategy, numShards)
}

func testAccMultiRegionClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccMultiRegionClusterConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
