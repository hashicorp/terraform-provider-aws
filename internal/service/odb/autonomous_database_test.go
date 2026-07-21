// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAccAutonomousDatabaseAdminPasswordEnv = "TF_VAR_odb_test_admin_password"
	testAccAutonomousDatabaseNetworkIDEnv     = "TF_VAR_odb_test_network_id"
)

func TestAccODBAutonomousDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database1, databaseAfterTagUpdate, databaseAfterMutableUpdate odbtypes.AutonomousDatabase
	resourceName := "aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	config := testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "test")
	tagUpdatedConfig := testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "updated")
	updatedConfig := testAccAutonomousDatabaseConfigBasic(displayName+"updated", dbName, 4, "AL32UTF8", "updated")
	replacementConfig := testAccAutonomousDatabaseConfigBasic(displayName+"updated", dbName, 4, "UTF8", "updated")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabasePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database1),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "data_storage_size_in_tbs", "1"),
					resource.TestCheckResourceAttr(resourceName, "db_name", dbName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestMatchResourceAttr(resourceName, names.AttrStatus, regexache.MustCompile(`^(AVAILABLE|AVAILABLE_NEEDS_ATTENTION|STOPPED|STANDBY)$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"admin_password_wo", "admin_password_wo_version", "source", "source_configuration", "transportable_tablespace"},
			},
			{
				Config: tagUpdatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseAfterTagUpdate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "updated"),
					func(*terraform.State) error {
						if aws.ToString(database1.AutonomousDatabaseId) != aws.ToString(databaseAfterTagUpdate.AutonomousDatabaseId) {
							return errors.New("Autonomous Database was replaced during a tag update")
						}
						return nil
					},
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseAfterMutableUpdate),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName+"updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "updated"),
					func(*terraform.State) error {
						if aws.ToString(database1.AutonomousDatabaseId) != aws.ToString(databaseAfterMutableUpdate.AutonomousDatabaseId) {
							return errors.New("Autonomous Database was replaced during an in-place update")
						}
						return nil
					},
				),
			},
			{
				Config:             replacementConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccODBAutonomousDatabase_allArguments(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database odbtypes.AutonomousDatabase
	resourceName := "aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabasePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutonomousDatabaseConfigAllArguments(displayName, dbName, acctest.DefaultEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database),
					resource.TestCheckResourceAttr(resourceName, "admin_password_wo_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowlisted_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowlisted_ips.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "autonomous_maintenance_schedule_type", "REGULAR"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period_in_days", "15"),
					resource.TestCheckResourceAttr(resourceName, "character_set", "AL32UTF8"),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "customer_contacts_to_send_to_oci.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customer_contacts_to_send_to_oci.0.email", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "data_storage_size_in_tbs", "1"),
					resource.TestCheckResourceAttr(resourceName, "database_edition", "ENTERPRISE_EDITION"),
					resource.TestCheckResourceAttr(resourceName, "db_name", dbName),
					resource.TestCheckResourceAttr(resourceName, "db_workload", "OLTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName),
					resource.TestCheckResourceAttr(resourceName, "encryption_key_provider", "ORACLE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "is_auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "is_auto_scaling_for_storage_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "is_backup_retention_locked", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_local_data_guard_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_mtls_connection_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "license_model", "BRING_YOUR_OWN_LICENSE"),
					resource.TestCheckResourceAttr(resourceName, "long_term_backup_schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "long_term_backup_schedule.0.is_disabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ncharacter_set", "AL16UTF16"),
					resource.TestCheckResourceAttrSet(resourceName, "odb_network_id"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_operations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_operations.0.day_of_week", "MONDAY"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_operations.0.scheduled_start_time", "08:00"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_operations.0.scheduled_stop_time", "18:00"),
					resource.TestCheckResourceAttr(resourceName, "source", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", displayName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func TestAccODBAutonomousDatabase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database odbtypes.AutonomousDatabase
	resourceName := "aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabasePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.ResourceAutonomousDatabase, resourceName),
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

func TestAccODBAutonomousDatabase_validation(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `resource "aws_odb_autonomous_database" "test" { db_name = "invalid-name" }`,
				ExpectError: regexache.MustCompile(`must start with a letter and contain only alphanumeric\s+characters`),
			},
			{
				Config:      `resource "aws_odb_autonomous_database" "test" { source = "INVALID" }`,
				ExpectError: regexache.MustCompile("Invalid String Enum Value"),
			},
		},
	})
}

func testAccAutonomousDatabasePreCheck(ctx context.Context, t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseAdminPasswordEnv)
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseNetworkIDEnv)
	testAccAutonomousDatabaseServicePreCheck(ctx, t)
}

func testAccAutonomousDatabaseServicePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
	_, err := conn.ListAutonomousDatabases(ctx, &odb.ListAutonomousDatabasesInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAutonomousDatabaseExists(ctx context.Context, t *testing.T, name string, database *odbtypes.AutonomousDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabase, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabase, name, errors.New("ID not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		found, err := tfodb.FindAutonomousDatabaseByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAutonomousDatabase, rs.Primary.ID, err)
		}

		*database = *found
		return nil
	}
}

func testAccCheckAutonomousDatabaseDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_autonomous_database" {
				continue
			}

			_, err := tfodb.FindAutonomousDatabaseByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAutonomousDatabase, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAutonomousDatabase, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccAutonomousDatabaseConfigPrerequisites() string {
	return `
variable "odb_test_admin_password" {
  type      = string
  sensitive = true
}

variable "odb_test_network_id" {
  type = string
}
`
}

func testAccAutonomousDatabaseConfigBasic(displayName, dbName string, computeCount float64, characterSet, environment string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigPrerequisites(),
		fmt.Sprintf(`
resource "aws_odb_autonomous_database" "test" {
  admin_password_wo         = var.odb_test_admin_password
  admin_password_wo_version = 1
  character_set             = %[1]q
  compute_count             = %[2]g
  data_storage_size_in_tbs  = 1
  db_name                   = %[3]q
  db_workload               = "OLTP"
  display_name              = %[4]q
  license_model             = "LICENSE_INCLUDED"
  odb_network_id            = var.odb_test_network_id
  source                    = "NONE"

  tags = {
    Environment = %[5]q
  }
}
`, characterSet, computeCount, dbName, displayName, environment),
	)
}

func testAccAutonomousDatabaseConfigAllArguments(displayName, dbName, emailAddress string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigPrerequisites(),
		fmt.Sprintf(`
resource "aws_odb_autonomous_database" "test" {
  admin_password_wo                    = var.odb_test_admin_password
  admin_password_wo_version            = 1
  allowlisted_ips                      = ["10.0.0.0/8"]
  autonomous_maintenance_schedule_type = "REGULAR"
  backup_retention_period_in_days      = 15
  character_set                        = "AL32UTF8"
  compute_count                        = 2
  data_storage_size_in_tbs             = 1
  database_edition                     = "ENTERPRISE_EDITION"
  db_name                              = %[1]q
  db_workload                          = "OLTP"
  display_name                         = %[2]q
  encryption_key_provider              = "ORACLE_MANAGED"
  is_auto_scaling_enabled              = true
  is_auto_scaling_for_storage_enabled  = true
  is_backup_retention_locked           = false
  is_local_data_guard_enabled          = false
  is_mtls_connection_required          = true
  license_model                        = "BRING_YOUR_OWN_LICENSE"
  ncharacter_set                       = "AL16UTF16"
  odb_network_id                       = var.odb_test_network_id
  source                               = "NONE"

  customer_contacts_to_send_to_oci {
    email = %[3]q
  }

  long_term_backup_schedule {
    is_disabled = true
  }

  scheduled_operations {
    day_of_week          = "MONDAY"
    scheduled_start_time = "08:00"
    scheduled_stop_time  = "18:00"
  }

  tags = {
    Environment = "test"
    Name        = %[2]q
  }
}
`, dbName, displayName, emailAddress),
	)
}
