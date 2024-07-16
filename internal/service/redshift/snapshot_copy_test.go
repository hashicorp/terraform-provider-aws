// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftSnapshotCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap types.ClusterSnapshotCopyStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_copy.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &snap),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "7"),
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

func TestAccRedshiftSnapshotCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap types.ClusterSnapshotCopyStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &snap),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceSnapshotCopy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSnapshotCopy_disappears_Cluster(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap types.ClusterSnapshotCopyStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_copy.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &snap),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSnapshotCopy_retentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap types.ClusterSnapshotCopyStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_copy.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_retentionPeriod(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &snap),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, acctest.Ct10),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotCopyConfig_retentionPeriod(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &snap),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "20"),
				),
			},
		},
	})
}

func testAccCheckSnapshotCopyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_snapshot_copy" {
				continue
			}

			_, err := tfredshift.FindSnapshotCopyByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Redshift, create.ErrActionCheckingDestroyed, tfredshift.ResNameSnapshotCopy, rs.Primary.ID, err)
			}

			return create.Error(names.Redshift, create.ErrActionCheckingDestroyed, tfredshift.ResNameSnapshotCopy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSnapshotCopyExists(ctx context.Context, name string, snap *types.ClusterSnapshotCopyStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameSnapshotCopy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameSnapshotCopy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)
		out, err := tfredshift.FindSnapshotCopyByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Redshift, create.ErrActionCheckingExistence, tfredshift.ResNameSnapshotCopy, rs.Primary.ID, err)
		}

		*snap = *out

		return nil
	}
}

func testAccSnapshotCopyConfigBase(rName string) string {
	return acctest.ConfigCompose(
		// "InvalidVPCNetworkStateFault: The requested AZ us-west-2a is not a valid AZ."
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  multi_az                            = false
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccSnapshotCopyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotCopyConfigBase(rName),
		fmt.Sprintf(`
resource "aws_redshift_snapshot_copy" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  destination_region = %[2]q
}
`, rName, acctest.AlternateRegion()))
}

func testAccSnapshotCopyConfig_retentionPeriod(rName string, retentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccSnapshotCopyConfigBase(rName),
		fmt.Sprintf(`
resource "aws_redshift_snapshot_copy" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  destination_region = %[2]q
  retention_period   = %[3]d
}
`, rName, acctest.AlternateRegion(), retentionPeriod))
}
