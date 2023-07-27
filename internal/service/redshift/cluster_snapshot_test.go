// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"regexp"
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

func TestAccRedshiftClusterSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.Snapshot
	resourceName := "aws_redshift_cluster_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift", regexp.MustCompile(`snapshot:.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account"),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "manual_snapshot_retention_period", "-1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterSnapshotConfig_retention(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift", regexp.MustCompile(`snapshot:.+`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "manual_snapshot_retention_period", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRedshiftClusterSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.Snapshot
	resourceName := "aws_redshift_cluster_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
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
				Config: testAccClusterSnapshotConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftClusterSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.Snapshot
	resourceName := "aws_redshift_cluster_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceClusterSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterSnapshotExists(ctx context.Context, n string, v *redshift.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Snapshot is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		out, err := tfredshift.FindClusterSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckClusterSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster_snapshot" {
				continue
			}

			_, err := tfredshift.FindClusterSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Cluster Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q
}
`, rName))
}

func testAccClusterSnapshotConfig_retention(rName string, retention int) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier               = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier              = %[1]q
  manual_snapshot_retention_period = %[2]d
}
`, rName, retention))
}

func testAccClusterSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterSnapshotConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
