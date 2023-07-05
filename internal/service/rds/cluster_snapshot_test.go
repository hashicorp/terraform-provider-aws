// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSClusterSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterSnapshot rds.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "db_cluster_snapshot_arn", "rds", regexp.MustCompile(fmt.Sprintf("cluster-snapshot:%s$", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, "port"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccRDSClusterSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterSnapshot rds.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccClusterSnapshotConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckClusterSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_cluster_snapshot" {
				continue
			}

			_, err := tfrds.FindDBClusterSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterSnapshotExists(ctx context.Context, n string, v *rds.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Cluster Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindDBClusterSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterSnapshotConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = %[1]q
  db_subnet_group_name            = aws_db_subnet_group.test.name
  database_name                   = "test"
  engine                          = "aurora-mysql"
  master_username                 = "tfacctest"
  master_password                 = "avoid-plaintext-passwords"
  db_cluster_parameter_group_name = "default.aurora-mysql5.7"
  skip_final_snapshot             = true
}
`, rName))
}

func testAccClusterSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccClusterSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterSnapshotConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
