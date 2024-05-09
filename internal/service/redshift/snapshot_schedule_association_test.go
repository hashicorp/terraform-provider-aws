// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftSnapshotScheduleAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"
	snapshotScheduleResourceName := "aws_redshift_snapshot_schedule.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "schedule_identifier", snapshotScheduleResourceName, names.AttrID),
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

func TestAccRedshiftSnapshotScheduleAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceSnapshotScheduleAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSnapshotScheduleAssociation_disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotScheduleAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_snapshot_schedule_association" {
				continue
			}

			clusterIdentifier, scheduleIdentifier, err := tfredshift.SnapshotScheduleAssociationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

			_, err = tfredshift.FindSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Snapshot Schedule Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotScheduleAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Snapshot Schedule Association ID is set")
		}

		clusterIdentifier, scheduleIdentifier, err := tfredshift.SnapshotScheduleAssociationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		_, err = tfredshift.FindSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

		return err
	}
}

func testAccSnapshotScheduleAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), testAccSnapshotScheduleConfig_basic(rName), `
resource "aws_redshift_snapshot_schedule_association" "test" {
  schedule_identifier = aws_redshift_snapshot_schedule.test.id
  cluster_identifier  = aws_redshift_cluster.test.id
}
`)
}
