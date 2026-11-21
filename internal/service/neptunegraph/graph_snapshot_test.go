// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfneptunegraph "github.com/hashicorp/terraform-provider-aws/internal/service/neptunegraph"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphGraphSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graphsnapshot neptunegraph.GetGraphSnapshotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
				Config: testAccGraphSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphSnapshotExists(ctx, t, resourceName, &graphsnapshot),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "graph_identifier", "aws_neptunegraph_graph.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_graph_id", "aws_neptunegraph_graph.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_create_time"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "neptune-graph", regexache.MustCompile(`graph-snapshot/.+$`)),
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

func TestAccNeptuneGraphGraphSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graphsnapshot neptunegraph.GetGraphSnapshotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGraphSnapshotExists(ctx, t, resourceName, &graphsnapshot),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfneptunegraph.ResourceGraphSnapshot, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNeptuneGraphGraphSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var graphsnapshot neptunegraph.GetGraphSnapshotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_graph_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphSnapshotExists(ctx, t, resourceName, &graphsnapshot),
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
				Config: testAccGraphSnapshotConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphSnapshotExists(ctx, t, resourceName, &graphsnapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphSnapshotConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphSnapshotExists(ctx, t, resourceName, &graphsnapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckGraphSnapshotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptunegraph_graph_snapshot" {
				continue
			}

			_, err := tfneptunegraph.FindGraphSnapshotByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Graph Graph Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGraphSnapshotExists(ctx context.Context, t *testing.T, n string, v *neptunegraph.GetGraphSnapshotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		output, err := tfneptunegraph.FindGraphSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGraphSnapshotConfig_basic(rName string) string {
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
`, rName)
}

func testAccGraphSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGraphSnapshotConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
