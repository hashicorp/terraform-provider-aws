package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMemoryDBSnapshot_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "snapshot/"+rName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.description", "aws_memorydb_cluster.test", "description"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.engine_version", "aws_memorydb_cluster.test", "engine_version"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.maintenance_window", "aws_memorydb_cluster.test", "maintenance_window"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.name", "aws_memorydb_cluster.test", "name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.node_type", "aws_memorydb_cluster.test", "node_type"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.num_shards", "aws_memorydb_cluster.test", "num_shards"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.parameter_group_name", "aws_memorydb_cluster.test", "parameter_group_name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.port", "aws_memorydb_cluster.test", "port"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.snapshot_retention_limit", "aws_memorydb_cluster.test", "snapshot_retention_limit"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.snapshot_window", "aws_memorydb_cluster.test", "snapshot_window"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.subnet_group_name", "aws_memorydb_cluster.test", "subnet_group_name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "cluster_configuration.0.vpc_id", "aws_memorydb_subnet_group.test", "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "source", "manual"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBSnapshot_nameGenerated(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_noName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSnapshot_namePrefix(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_namePrefix(rName, "tftest-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tftest-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tftest-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSnapshot_create_withKMS(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags2(rName, "Key1", "value1", "Key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags1(rName, "Key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
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
					testAccCheckSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func testAccCheckSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_snapshot" {
			continue
		}

		_, err := tfmemorydb.FindSnapshotByName(context.Background(), conn, rs.Primary.Attributes["name"])

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

func testAccCheckSnapshotExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindSnapshotByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSnapshotConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  subnet_ids = aws_subnet.test.*.id
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
