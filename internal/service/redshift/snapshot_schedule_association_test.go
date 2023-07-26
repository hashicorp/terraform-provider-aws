// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftSnapshotScheduleAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"
	snapshotScheduleResourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "schedule_identifier", snapshotScheduleResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", clusterResourceName, "id"),
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
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
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
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
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

			conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

			_, _, err := tfredshift.FindScheduleAssociationById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Schedule Association %s still exists", rs.Primary.ID)
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
			return fmt.Errorf("No Redshift Cluster Snapshot Schedule Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		_, _, err := tfredshift.FindScheduleAssociationById(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccSnapshotScheduleAssociationConfig_basic(rName, definition string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), testAccSnapshotScheduleConfig_basic(rName, definition), `
resource "aws_redshift_snapshot_schedule_association" "test" {
  schedule_identifier = aws_redshift_snapshot_schedule.default.id
  cluster_identifier  = aws_redshift_cluster.test.id
}
`)
}
