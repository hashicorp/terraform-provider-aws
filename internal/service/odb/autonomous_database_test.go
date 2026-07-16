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
	testAccAutonomousDatabaseAZIDEnv          = "TF_VAR_odb_test_availability_zone_id"
)

func TestAccODBAutonomousDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database1, database2 odbtypes.AutonomousDatabase
	resourceName := "aws_odb_autonomous_database.test"
	networkName := acctest.RandomWithPrefix(t, "tf-odb-net")
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	secondOctet := acctest.RandIntRange(t, 10, 200)

	config := testAccAutonomousDatabaseConfigBasic(networkName, displayName, dbName, secondOctet, 2, "AL32UTF8", "test")
	updatedConfig := testAccAutonomousDatabaseConfigBasic(networkName, displayName+"updated", dbName, secondOctet, 4, "AL32UTF8", "updated")
	replacementConfig := testAccAutonomousDatabaseConfigBasic(networkName, displayName+"updated", dbName, secondOctet, 4, "UTF8", "updated")

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
				Config: updatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database2),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName+"updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "updated"),
					func(*terraform.State) error {
						if aws.ToString(database1.AutonomousDatabaseId) != aws.ToString(database2.AutonomousDatabaseId) {
							return errors.New("Autonomous Database was replaced during an in-place update")
						}
						return nil
					},
				),
			},
			{
				Config:   replacementConfig,
				PlanOnly: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
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
	networkName := acctest.RandomWithPrefix(t, "tf-odb-net")
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	secondOctet := acctest.RandIntRange(t, 10, 200)

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
				Config: testAccAutonomousDatabaseConfigAllArguments(networkName, displayName, dbName, secondOctet),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database),
					resource.TestCheckResourceAttr(resourceName, "autonomous_maintenance_schedule_type", "REGULAR"),
					resource.TestCheckResourceAttr(resourceName, "customer_contacts_to_send_to_oci.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_operations.0.day_of_week", "MONDAY"),
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
	networkName := acctest.RandomWithPrefix(t, "tf-odb-net")
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	secondOctet := acctest.RandIntRange(t, 10, 200)

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
				Config: testAccAutonomousDatabaseConfigBasic(networkName, displayName, dbName, secondOctet, 2, "AL32UTF8", "test"),
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
				ExpectError: regexache.MustCompile("must start with a letter and contain only alphanumeric characters"),
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
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseAZIDEnv)
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

func testAccAutonomousDatabaseConfigPrerequisites(networkName string, secondOctet int) string {
	return fmt.Sprintf(`
variable "odb_test_admin_password" {
  type      = string
  sensitive = true
}

variable "odb_test_availability_zone_id" {
  type = string
}

resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = var.odb_test_availability_zone_id
  client_subnet_cidr   = "10.%[2]d.0.0/24"
  backup_subnet_cidr   = "10.%[2]d.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}
`, networkName, secondOctet)
}

func testAccAutonomousDatabaseConfigBasic(networkName, displayName, dbName string, secondOctet int, computeCount float64, characterSet, environment string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigPrerequisites(networkName, secondOctet),
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
  odb_network_id            = aws_odb_network.test.id
  source                    = "NONE"

  tags = {
    Environment = %[5]q
  }
}
`, characterSet, computeCount, dbName, displayName, environment),
	)
}

func testAccAutonomousDatabaseConfigAllArguments(networkName, displayName, dbName string, secondOctet int) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigPrerequisites(networkName, secondOctet),
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
  license_model                        = "LICENSE_INCLUDED"
  ncharacter_set                       = "AL16UTF16"
  odb_network_id                       = aws_odb_network.test.id
  source                               = "NONE"

  customer_contacts_to_send_to_oci {
    email = "terraform-odb@example.com"
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
`, dbName, displayName),
	)
}
