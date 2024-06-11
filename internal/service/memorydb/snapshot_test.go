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

func TestAccMemoryDBSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "memorydb", "snapshot/"+rName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.description", "aws_memorydb_cluster.test", names.AttrDescription),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.engine_version", "aws_memorydb_cluster.test", names.AttrEngineVersion),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.maintenance_window", "aws_memorydb_cluster.test", "maintenance_window"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.name", "aws_memorydb_cluster.test", names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.node_type", "aws_memorydb_cluster.test", "node_type"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.num_shards", "aws_memorydb_cluster.test", "num_shards"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.parameter_group_name", "aws_memorydb_cluster.test", names.AttrParameterGroupName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.port", "aws_memorydb_cluster.test", names.AttrPort),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.snapshot_retention_limit", "aws_memorydb_cluster.test", "snapshot_retention_limit"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.snapshot_window", "aws_memorydb_cluster.test", "snapshot_window"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.subnet_group_name", "aws_memorydb_cluster.test", "subnet_group_name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.vpc_id", "aws_memorydb_subnet_group.test", names.AttrVPCID),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, "manual"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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

func TestAccMemoryDBSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmemorydb.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBSnapshot_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_noName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSnapshot_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_namePrefix(rName, "tftest-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tftest-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tftest-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSnapshot_create_withKMS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
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

func TestAccMemoryDBSnapshot_update_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags2(rName, "Key1", acctest.CtValue1, "Key2", acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags1(rName, "Key1", acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
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

func testAccCheckSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_memorydb_snapshot" {
				continue
			}

			_, err := tfmemorydb.FindSnapshotByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MemoryDB Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn(ctx)

		_, err := tfmemorydb.FindSnapshotByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccSnapshotConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_memorydb_cluster" "test" {
  acl_name                 = "open-access"
  name                     = %[1]q
  node_type                = "db.t4g.small"
  num_replicas_per_shard   = 0
  num_shards               = 1
  security_group_ids       = [aws_security_group.test.id]
  snapshot_retention_limit = 0
  subnet_group_name        = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  name         = %[1]q

  tags = {
    Test = "test"
  }
}
`, rName),
	)
}

func testAccSnapshotConfig_kms(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  kms_key_arn  = aws_kms_key.test.arn
  name         = %[1]q
}
`, rName),
	)
}

func testAccSnapshotConfig_noName(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
}
`,
	)
}

func testAccSnapshotConfig_namePrefix(rName, prefix string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  name_prefix  = %[1]q

  tags = {
    Test = "test"
  }
}
`, prefix),
	)
}

func testAccSnapshotConfig_tags0(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  name         = %[1]q
}
`, rName),
	)
}

func testAccSnapshotConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  name         = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value),
	)
}

func testAccSnapshotConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  name         = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value),
	)
}
