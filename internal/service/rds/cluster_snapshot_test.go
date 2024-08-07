// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterSnapshot types.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zones.#"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "db_cluster_snapshot_arn", "rds", regexache.MustCompile(fmt.Sprintf("cluster-snapshot:%s$", rName))),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snapshot_type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "source_db_cluster_snapshot_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "available"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrVPCID, regexache.MustCompile(`^vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccClusterSnapshotConfig_sharedAccounts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "shared_accounts.*", "all"),
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
	var dbClusterSnapshot types.DBClusterSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccClusterSnapshotConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterSnapshotConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &dbClusterSnapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckClusterSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

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

func testAccCheckClusterSnapshotExists(ctx context.Context, n string, v *types.DBClusterSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBClusterSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterSnapshotConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  database_name        = "test"
  engine               = "aurora-mysql"
  master_username      = "tfacctest"
  master_password      = "avoid-plaintext-passwords"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccClusterSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccClusterSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotConfig_base(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccClusterSnapshotConfig_base(rName),
		fmt.Sprintf(`
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

func testAccClusterSnapshotConfig_sharedAccounts(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q
  shared_accounts                = ["all"]
}
`, rName))
}
