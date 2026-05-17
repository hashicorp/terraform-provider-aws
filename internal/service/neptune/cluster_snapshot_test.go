// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneClusterSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterSnapshot awstypes.DBClusterSnapshot
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, t, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "db_cluster_snapshot_arn", "rds", fmt.Sprintf("cluster-snapshot:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "available"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrVPCID, regexache.MustCompile(`^vpc-.+`)),
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

func TestAccNeptuneClusterSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterSnapshot awstypes.DBClusterSnapshot
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, t, resourceName, &dbClusterSnapshot),
					acctest.CheckSDKResourceDisappears(ctx, t, tfneptune.ResourceClusterSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterSnapshotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NeptuneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster_snapshot" {
				continue
			}

			_, err := tfneptune.FindClusterSnapshotByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterSnapshotExists(ctx context.Context, t *testing.T, n string, v *awstypes.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Cluster Snapshot ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).NeptuneClient(ctx)

		output, err := tfneptune.FindClusterSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterSnapshotConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  skip_final_snapshot = true

  neptune_cluster_parameter_group_name = "default.neptune1.3"
}

resource "aws_neptune_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_neptune_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
}
`, rName)
}
