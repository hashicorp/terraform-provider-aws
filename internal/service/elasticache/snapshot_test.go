package elasticache_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheSnapshot_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshot elasticache.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`snapshot:+.`)),
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

func TestAccElastiCacheSnapshot_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshot elasticache.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSnapshotConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccElastiCacheSnapshot_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snapshot elasticache.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_snapshot" {
			continue
		}

		resp, err := conn.DescribeSnapshots(&elasticache.DescribeSnapshotsInput{
			SnapshotName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.Snapshots) != 0 && aws.StringValue(resp.Snapshots[0].SnapshotName) == rs.Primary.ID {
				return create.Error(names.ElastiCache, create.ErrActionCheckingDestroyed, tfelasticache.ResNameSnapshot, rs.Primary.ID, errors.New("not destroyed"))
			}
		}
	}

	return nil
}

func testAccCheckSnapshotExists(name string, snapshot *elasticache.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ElastiCache, create.ErrActionCheckingExistence, tfelasticache.ResNameSnapshot, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ElastiCache, create.ErrActionCheckingExistence, tfelasticache.ResNameSnapshot, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		resp, err := conn.DescribeSnapshots(&elasticache.DescribeSnapshotsInput{
			SnapshotName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ElastiCache, create.ErrActionCheckingExistence, tfelasticache.ResNameSnapshot, rs.Primary.ID, err)
		}

		*snapshot = *resp.Snapshots[0]

		return nil
	}
}

func testAccSnapshotBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_elasticache_snapshot" "test" {
  cluster_id    = aws_elasticache_cluster.test.cluster_id
  snapshot_name = %[1]q
}
`, rName))
}

func testAccSnapshotConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccSnapshotBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_elasticache_snapshot" "test" {
  cluster_id    = aws_elasticache_cluster.test.cluster_id
  snapshot_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccSnapshotConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccSnapshotBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_elasticache_snapshot" "test" {
  cluster_id    = aws_elasticache_cluster.test.cluster_id
  snapshot_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
