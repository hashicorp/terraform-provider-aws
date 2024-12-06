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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNameSuffix),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttrSet(resourceName, "num_shards"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNameSuffix),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
				Config: testAccMultiRegionClusterConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Also managed by Terraform"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_clusters(t *testing.T) {
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
				Config: testAccMultiRegionClusterConfig_clusters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "clusters.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "clusters.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "clusters.0.region"),
					resource.TestCheckResourceAttrSet(resourceName, "clusters.0.status"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
  name_suffix = %[1]q
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
  name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
}
`, rName)
}

func testAccMultiRegionClusterConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  description = "Also managed by Terraform"

  tags = {
    Test = "test"
  }
}
`, rName)
}

func testAccMultiRegionClusterConfig_tlsEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  name_suffix = %[1]q
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
  name_suffix = %[1]q
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
  name_suffix    = %[1]q
  node_type      = "db.r7g.xlarge"
  engine         = %[2]q
  engine_version = %[3]q

  tags = {
    Test = "test"
  }
}
`, rName, engine, engineVersion)
}

func testAccMultiRegionClusterConfig_clusters(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
		testAccClusterConfig_baseUserAndACL(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_memorydb_cluster" "test" {
  acl_name                   = aws_memorydb_acl.test.id
  auto_minor_version_upgrade = false
  name                       = %[1]q
  node_type                  = "db.r7g.xlarge"
  num_shards                 = 2
  security_group_ids         = [aws_security_group.test.id]
  snapshot_retention_limit   = 7
  subnet_group_name          = aws_memorydb_subnet_group.test.id

  multi_region_cluster_name = aws_memorydb_multi_region_cluster.test.name

  tags = {
    Test = "test"
  }
}

resource "aws_memorydb_multi_region_cluster" "test" {
  name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"
  num_shards  = 2

  tags = {
    Test = "test"
  }
}
`, rName),
	)
}

func testAccMultiRegionClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  name_suffix = %[1]q
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
  name_suffix = %[1]q
  node_type   = "db.r7g.xlarge"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
