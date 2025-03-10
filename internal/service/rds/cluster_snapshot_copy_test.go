// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

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

func TestAccRDSClusterSnapshotCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccRDSClusterSnapshotCopy_share(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_share(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "shared_accounts.*", "all"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", "0"),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotCopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
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
				Config: testAccClusterSnapshotCopyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterSnapshotCopyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrds.ResourceClusterSnapshotCopy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterSnapshotCopy_destinationRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_destinationRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"destination_region"},
			},
		},
	})
}

func TestAccRDSClusterSnapshotCopy_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBClusterSnapshot
	resourceName := "aws_rds_cluster_snapshot_copy.test"
	keyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotCopyConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckClusterSnapshotCopyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_snapshot_copy" {
				continue
			}

			_, err := tfrds.FindDBClusterSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Cluster Snapshot Copy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterSnapshotCopyExists(ctx context.Context, n string, v *types.DBClusterSnapshot) resource.TestCheckFunc {
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

func testAccClusterSnapshotCopyConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.cluster_identifier
  db_cluster_snapshot_identifier = "%[1]s-source"
}`, rName)
}

func testAccClusterSnapshotCopyConfig_encryptedBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "encrypted" {
  cluster_identifier  = %[1]q
  database_name       = "test"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
  storage_encrypted   = true
}

resource "aws_db_cluster_snapshot" "encrypted" {
  db_cluster_identifier          = aws_rds_cluster.encrypted.cluster_identifier
  db_cluster_snapshot_identifier = "%[1]s-source"
}`, rName)
}

func testAccClusterSnapshotCopyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"
}`, rName))
}

func testAccClusterSnapshotCopyConfig_tags1(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
  }
}`, rName, tagKey, tagValue))
}

func testAccClusterSnapshotCopyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterSnapshotCopyConfig_share(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"
  shared_accounts                       = ["all"]
}
`, rName))
}

func testAccClusterSnapshotCopyConfig_destinationRegion(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"
  destination_region                    = %[2]q
}`, rName, acctest.AlternateRegion()))
}

func testAccClusterSnapshotCopyConfig_kms(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterSnapshotCopyConfig_encryptedBase(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "test"
}

resource "aws_rds_cluster_snapshot_copy" "test" {
  source_db_cluster_snapshot_identifier = aws_db_cluster_snapshot.encrypted.db_cluster_snapshot_arn
  target_db_cluster_snapshot_identifier = "%[1]s-target"
  kms_key_id                            = aws_kms_key.test.arn
}`, rName))
}
