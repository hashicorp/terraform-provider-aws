// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphGraphSnapshotsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_neptunegraph_graph_snapshots.test"
	resourceName := "aws_neptunegraph_graph_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphSnapshotsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "graph_snapshots.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "graph_snapshots.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "graph_snapshots.0.name", resourceName, "snapshot_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "graph_snapshots.0.source_graph_id", resourceName, "source_graph_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "graph_snapshots.0.arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "graph_snapshots.0.status"),
				),
			},
		},
	})
}

func testAccGraphSnapshotsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = false
}

resource "aws_neptunegraph_graph_snapshot" "test" {
  graph_identifier = aws_neptunegraph_graph.test.id
  snapshot_name    = %[1]q
}

data "aws_neptunegraph_graph_snapshots" "test" {
  graph_identifier = aws_neptunegraph_graph.test.id

  depends_on = [aws_neptunegraph_graph_snapshot.test]
}
`, rName)
}
