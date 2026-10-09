// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheSnapshotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Create the cache cluster first so a snapshot can be taken from it.
				Config: testAccSnapshotDataSourceConfig_cluster(rName),
			},
			{
				// The provider has no aws_elasticache_snapshot resource, so create the
				// snapshot out-of-band before reading it through the data source.
				PreConfig: func() {
					testAccCreateSnapshot(ctx, t, rName, rName)
				},
				Config: testAccSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_id", rName),
					resource.TestCheckResourceAttr(dataSourceName, "snapshot_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "elasticache", regexache.MustCompile("snapshot:"+rName+"$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttrSet(dataSourceName, "node_type"),
					resource.TestCheckResourceAttr(dataSourceName, "snapshot_source", "manual"),
					resource.TestCheckResourceAttrSet(dataSourceName, "node_snapshots.#"),
				),
			},
		},
	})
}

// testAccCreateSnapshot creates a snapshot from the cluster via the AWS API, waits for it to become available, and registers cleanup to delete it.
func testAccCreateSnapshot(ctx context.Context, t *testing.T, clusterID, snapshotName string) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

	_, err := conn.CreateSnapshot(ctx, &elasticache.CreateSnapshotInput{
		CacheClusterId: aws.String(clusterID),
		SnapshotName:   aws.String(snapshotName),
	})
	if err != nil {
		t.Fatalf("creating ElastiCache Snapshot (%s): %s", snapshotName, err)
	}

	t.Cleanup(func() {
		// The test context is canceled by the time cleanup runs, so use a non-canceled context.
		_, err := conn.DeleteSnapshot(context.WithoutCancel(ctx), &elasticache.DeleteSnapshotInput{
			SnapshotName: aws.String(snapshotName),
		})
		if err != nil && !errs.IsA[*awstypes.SnapshotNotFoundFault](err) {
			t.Logf("deleting ElastiCache Snapshot (%s): %s", snapshotName, err)
		}
	})

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Timeout:    40 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
		Refresh: func(ctx context.Context) (any, string, error) {
			output, err := conn.DescribeSnapshots(ctx, &elasticache.DescribeSnapshotsInput{
				SnapshotName: aws.String(snapshotName),
			})
			if err != nil {
				return nil, "", err
			}
			if len(output.Snapshots) == 0 {
				return nil, "", nil
			}

			snapshot := output.Snapshots[0]
			return snapshot, aws.ToString(snapshot.SnapshotStatus), nil
		},
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		t.Fatalf("waiting for ElastiCache Snapshot (%s) to become available: %s", snapshotName, err)
	}
}

func testAccSnapshotDataSourceConfig_cluster(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotDataSourceConfig_cluster(rName), `
data "aws_elasticache_snapshot" "test" {
  cluster_id  = aws_elasticache_cluster.test.cluster_id
  most_recent = true
}
`)
}
