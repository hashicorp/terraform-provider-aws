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

func TestAccRDSSnapshotCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccRDSSnapshotCopy_share(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_share(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "shared_accounts.*", "all"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSSnapshotCopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotCopyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSnapshotCopyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSSnapshotCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceSnapshotCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSSnapshotCopy_destinationRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBSnapshot
	resourceName := "aws_db_snapshot_copy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSnapshotCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyConfig_destinationRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotCopyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "destination_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIOPS),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttrSet(resourceName, "license_model"),
					resource.TestCheckResourceAttrSet(resourceName, "option_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPort),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStorageType),
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

func testAccCheckSnapshotCopyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_snapshot_copy" {
				continue
			}

			_, err := tfrds.FindDBSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Snapshot Copy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotCopyExists(ctx context.Context, n string, v *types.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBSnapshotByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSnapshotCopyConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  identifier              = %[1]q
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  maintenance_window      = "Fri:09:00-Fri:09:30"
  backup_retention_period = 0
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = "%[1]s-source"
}`, rName)
}

func testAccSnapshotCopyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotCopyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"
}`, rName))
}

func testAccSnapshotCopyConfig_tags1(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(testAccSnapshotCopyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
  }
}`, rName, tagKey, tagValue))
}

func testAccSnapshotCopyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccSnapshotCopyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSnapshotCopyConfig_share(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotCopyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"
  shared_accounts               = ["all"]
}
`, rName))
}

func testAccSnapshotCopyConfig_destinationRegion(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotCopyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-target"
  destination_region            = %[2]q
}`, rName, acctest.AlternateRegion()))
}
