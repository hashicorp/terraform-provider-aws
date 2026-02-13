// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMemoryDBMultiRegionCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "memorydb", "multiregioncluster/{multi_region_cluster_name}"),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_cluster_name"),
					resource.TestCheckResourceAttr(resourceName, "multi_region_cluster_name_suffix", rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, "multi_region_parameter_group_name"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "clusters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmemorydb.ResourceMultiRegionCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_description(rName, "Also managed by Terraform"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Also managed by Terraform"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_tlsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_tlsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
			{
				Config: testAccMultiRegionClusterConfig_tlsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_engine(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_engine(rName, "valkey"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_engineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_engineVersion(rName, "valkey", "7.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "7.3"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_updateStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_updateStrategy(rName, "coordinated", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "update_strategy", "coordinated"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_updateStrategy(rName, "uncoordinated", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "update_strategy", "uncoordinated"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "3"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
				ImportStateVerifyIgnore:              []string{"update_strategy"},
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_numShards(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_numShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
			{
				Config: testAccMultiRegionClusterConfig_numShards(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "3"),
				),
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_nodeType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_nodeType(rName, "db.r7g.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.xlarge"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
			{
				Config: testAccMultiRegionClusterConfig_nodeType(rName, "db.r7g.2xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.r7g.2xlarge"),
				),
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_parameterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_parameterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_parameter_group_name", "default.memorydb-valkey7.multiregion"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
		},
	})
}

func TestAccMemoryDBMultiRegionCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_memorydb_multi_region_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMultiRegionClusterImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "multi_region_cluster_name",
			},
			{
				Config: testAccMultiRegionClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMultiRegionClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionClusterExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckMultiRegionClusterExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		name := rs.Primary.Attributes["multi_region_cluster_name"]
		if name == "" {
			return fmt.Errorf("No MemoryDB Multi Region Cluster Name is set")
		}

		conn := acctest.ProviderMeta(ctx, t).MemoryDBClient(ctx)
		_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, name)

		return err
	}
}

func testAccCheckMultiRegionClusterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MemoryDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_memorydb_multi_region_cluster" {
				continue
			}

			name := rs.Primary.Attributes["multi_region_cluster_name"]

			_, err := tfmemorydb.FindMultiRegionClusterByName(ctx, conn, name)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MemoryDB Multi Region Cluster %s still exists", name)
		}

		return nil
	}
}

func testAccMultiRegionClusterImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["multi_region_cluster_name"], nil
	}
}

func testAccMultiRegionClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
}
`, rName)
}

// Sets `num_shards` to also test an update of the resource
func testAccMultiRegionClusterConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  description                      = %[2]q
}
`, rName, description)
}

func testAccMultiRegionClusterConfig_numShards(rName string, numShards int) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  description                      = "Also managed by Terraform"

  num_shards = %[2]d
}
`, rName, numShards)
}

func testAccMultiRegionClusterConfig_nodeType(rName string, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = %[2]q
  description                      = "Also managed by Terraform"
}
`, rName, nodeType)
}

func testAccMultiRegionClusterConfig_parameterGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  description                      = "Also managed by Terraform"

  multi_region_parameter_group_name = "default.memorydb-valkey7.multiregion"
}
`, rName)
}

func testAccMultiRegionClusterConfig_tlsEnabled(rName string, tlsEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  tls_enabled                      = %[2]t
}
`, rName, tlsEnabled)
}

func testAccMultiRegionClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  engine                           = %[2]q
}
`, rName, engine)
}

func testAccMultiRegionClusterConfig_engineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"
  engine                           = %[2]q
  engine_version                   = %[3]q
}
`, rName, engine, engineVersion)
}

// Sets `num_shards` to update the resource
func testAccMultiRegionClusterConfig_updateStrategy(rName, updateStrategy string, numShards int) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"

  update_strategy = %[2]q
  num_shards      = %[3]d
}
`, rName, updateStrategy, numShards)
}

func testAccMultiRegionClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_multi_region_cluster" "test" {
  multi_region_cluster_name_suffix = %[1]q
  node_type                        = "db.r7g.xlarge"

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
  node_type                        = "db.r7g.xlarge"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
