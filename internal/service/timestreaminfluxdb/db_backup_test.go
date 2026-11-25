// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamInfluxDBDBBackup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbBackup timestreaminfluxdb.GetDbBackupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_backup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBBackups(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBBackupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBBackupExists(ctx, t, resourceName, &dbBackup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestream-influxdb", regexache.MustCompile(`db-backup/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "db_resource_id", "aws_timestreaminfluxdb_db_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.DbBackupStatusCompleted)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.DbBackupTypeOnDemand)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBBackup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbBackup timestreaminfluxdb.GetDbBackupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_backup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBBackups(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBBackupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBBackupExists(ctx, t, resourceName, &dbBackup),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tftimestreaminfluxdb.ResourceDBBackup, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBBackup_retentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbBackup timestreaminfluxdb.GetDbBackupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_backup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBBackups(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBBackupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBBackupConfig_retentionDays(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBBackupExists(ctx, t, resourceName, &dbBackup),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_days"},
			},
		},
	})
}

func testAccCheckDBBackupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreaminfluxdb_db_backup" {
				continue
			}

			_, err := tftimestreaminfluxdb.FindDBBackupByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			return fmt.Errorf("Timestream InfluxDB DB Backup %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDBBackupExists(ctx context.Context, t *testing.T, n string, v *timestreaminfluxdb.GetDbBackupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)

		resp, err := tftimestreaminfluxdb.FindDBBackupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckDBBackups(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)

	input := &timestreaminfluxdb.ListDbBackupsInput{}
	_, err := conn.ListDbBackups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDBBackupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
}

resource "aws_timestreaminfluxdb_db_backup" "test" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.test.id
  name           = %[1]q
}
`, rName))
}

func testAccDBBackupConfig_retentionDays(rName string, retentionDays int) string {
	return acctest.ConfigCompose(testAccDBInstanceConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
}

resource "aws_timestreaminfluxdb_db_backup" "test" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.test.id
  name           = %[1]q
  retention_days = %[2]d
}
`, rName, retentionDays))
}
