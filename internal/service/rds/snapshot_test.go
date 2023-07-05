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

func TestAccRDSSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "db_snapshot_arn", "rds", regexp.MustCompile(`snapshot:.+`)),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", "0"),
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

func TestAccRDSSnapshot_share(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_share(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
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
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "shared_accounts.#", "0"),
				),
			},
		},
	})
}

func TestAccRDSSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	resourceName := "aws_db_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
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
				Config: testAccSnapshotConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSnapshotConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRDSSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSnapshotExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDBSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_snapshot" {
				continue
			}

			_, err := tfrds.FindDBSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDBSnapshotExists(ctx context.Context, n string, v *rds.DBSnapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS DB Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindDBSnapshotByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSnapshotConfig_base(rName string) string {
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
}`, rName)
}

func testAccSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccSnapshotConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccSnapshotConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccSnapshotConfig_share(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
  shared_accounts        = ["all"]
}
`, rName))
}
