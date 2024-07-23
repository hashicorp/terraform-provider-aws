// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstanceAutomatedBackupsReplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInstanceAutomatedBackupsReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(ctx, resourceName),
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

func TestAccRDSInstanceAutomatedBackupsReplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInstanceAutomatedBackupsReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceInstanceAutomatedBackupsReplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSInstanceAutomatedBackupsReplication_retentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInstanceAutomatedBackupsReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_retentionPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "14"),
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

func TestAccRDSInstanceAutomatedBackupsReplication_kmsEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance_automated_backups_replication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInstanceAutomatedBackupsReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAutomatedBackupsReplicationConfig_kmsEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceAutomatedBackupsReplicationExist(ctx, resourceName),
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

func testAccCheckInstanceAutomatedBackupsReplicationExist(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS instance automated backups replication ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		_, err := tfrds.FindDBInstanceAutomatedBackupByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckInstanceAutomatedBackupsReplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance_automated_backups_replication" {
				continue
			}

			_, err := tfrds.FindDBInstanceAutomatedBackupByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS instance automated backups replication %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceAutomatedBackupsReplicationConfig_base(rName string, storageEncrypted bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }

  provider = "awsalternate"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }

  provider = "awsalternate"
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block        = "10.1.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  provider = "awsalternate"
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }

  provider = "awsalternate"
}

data "aws_rds_engine_version" "default" {
  engine = "postgres"

  provider = "awsalternate"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[3]s]

  provider = "awsalternate"
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  identifier              = %[1]q
  engine                  = data.aws_rds_engine_version.default.engine
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  backup_retention_period = 7
  skip_final_snapshot     = true
  storage_encrypted       = %[2]t
  db_subnet_group_name    = aws_db_subnet_group.test.name

  provider = "awsalternate"
}
`, rName, storageEncrypted, mainInstanceClasses))
}

func testAccInstanceAutomatedBackupsReplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceAutomatedBackupsReplicationConfig_base(rName, false), `
resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
}
`)
}

func testAccInstanceAutomatedBackupsReplicationConfig_retentionPeriod(rName string) string {
	return acctest.ConfigCompose(testAccInstanceAutomatedBackupsReplicationConfig_base(rName, false), `
resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  retention_period       = 14
}
`)
}

func testAccInstanceAutomatedBackupsReplicationConfig_kmsEncrypted(rName string) string {
	return acctest.ConfigCompose(testAccInstanceAutomatedBackupsReplicationConfig_base(rName, true), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_db_instance_automated_backups_replication" "test" {
  source_db_instance_arn = aws_db_instance.test.arn
  kms_key_id             = aws_kms_key.test.arn
}
`, rName))
}
