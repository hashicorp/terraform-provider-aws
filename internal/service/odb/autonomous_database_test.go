// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAccAutonomousDatabaseAdminPasswordEnv              = "TF_VAR_odb_test_admin_password"
	testAccAutonomousDatabaseAdminPasswordSecretARNEnv     = "TF_VAR_odb_test_admin_password_secret_arn"
	testAccAutonomousDatabaseAdminPasswordSecretRoleARNEnv = "TF_VAR_odb_test_admin_password_secret_role_arn"
	testAccAutonomousDatabaseExternalIDTypeEnv             = "TF_VAR_odb_test_external_id_type"
	testAccAutonomousDatabaseNetworkIDEnv                  = "TF_VAR_odb_test_network_id"
)

func TestAccODBAutonomousDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database1, databaseAfterImport, databaseAfterTagUpdate, databaseAfterPasswordUpdate, databaseAfterMutableUpdate odbtypes.AutonomousDatabase
	bootstrapResourceName := "aws_odb_autonomous_database.bootstrap"
	resourceName := "aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	// Import cannot recover the password, so reconciliation must use a fresh value
	// to avoid the database password history policy. Keep each value stable between rotations.
	bootstrapVariables := config.Variables{
		"odb_test_admin_password": config.StringVariable("Create1#" + acctest.RandStringFromCharSet(t, 20, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")),
	}
	basicVariables := config.Variables{
		"odb_test_admin_password": config.StringVariable("Import2#" + acctest.RandStringFromCharSet(t, 20, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")),
	}
	updatedVariables := config.Variables{
		"odb_test_admin_password": config.StringVariable("Update3#" + acctest.RandStringFromCharSet(t, 20, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")),
	}
	basicConfig := testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "test")
	importVariables := config.Variables{
		"odb_test_admin_password": basicVariables["odb_test_admin_password"],
	}
	tagUpdatedConfig := testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "updated")
	updatedConfig := testAccAutonomousDatabaseConfigBasic(displayName+"updated", dbName, 4, "AL32UTF8", "updated")
	replacementConfig := testAccAutonomousDatabaseConfigBasic(displayName+"updated", dbName, 4, "UTF8", "updated")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_7_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseNetworkIDEnv)
			testAccAutonomousDatabaseServicePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:          testAccAutonomousDatabaseConfigBasicNamed("bootstrap", displayName, dbName, 2, "AL32UTF8", "test"),
				ConfigVariables: bootstrapVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, bootstrapResourceName, &database1),
					resource.TestCheckResourceAttr(bootstrapResourceName, "compute_count", "2"),
					resource.TestCheckResourceAttr(bootstrapResourceName, "data_storage_size_in_tbs", "1"),
					resource.TestCheckResourceAttr(bootstrapResourceName, "db_name", dbName),
					resource.TestCheckResourceAttr(bootstrapResourceName, "tags.Environment", "test"),
					resource.TestMatchResourceAttr(bootstrapResourceName, names.AttrStatus, regexache.MustCompile(`^(AVAILABLE|AVAILABLE_NEEDS_ATTENTION|STOPPED|STANDBY)$`)),
				),
			},
			{
				ResourceName:            bootstrapResourceName,
				ConfigVariables:         bootstrapVariables,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"admin_password", "admin_password_wo", "admin_password_wo_version", names.AttrSource, "source_configuration", "transportable_tablespace"},
			},
			{
				// The CLI import harness cannot persist a same-address re-import. Apply an import
				// into a new address while forgetting the bootstrap address without deleting the database.
				PreConfig: func() {
					t.Log("reconciling imported autonomous database with a fresh generated password")
					importVariables["import_id"] = config.StringVariable(aws.ToString(database1.AutonomousDatabaseId))
				},
				Config:          testAccAutonomousDatabaseConfigImport(basicConfig),
				ConfigVariables: importVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseAfterImport),
					func(s *terraform.State) error {
						if aws.ToString(database1.AutonomousDatabaseId) != aws.ToString(databaseAfterImport.AutonomousDatabaseId) {
							return errors.New("autonomous database was replaced during import reconciliation")
						}
						if _, ok := s.RootModule().Resources[bootstrapResourceName]; ok {
							return errors.New("bootstrap resource remains in state after import")
						}
						return nil
					},
				),
			},
			{
				Config:          basicConfig,
				ConfigVariables: basicVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config:          tagUpdatedConfig,
				ConfigVariables: basicVariables,
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
				PreConfig: func() {
					t.Log("rotating autonomous database password to a fresh generated value")
				},
				Config:          tagUpdatedConfig,
				ConfigVariables: updatedVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseAfterPasswordUpdate),
					resource.TestCheckResourceAttr(resourceName, "compute_count", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName),
					func(*terraform.State) error {
						if aws.ToString(database1.AutonomousDatabaseId) != aws.ToString(databaseAfterPasswordUpdate.AutonomousDatabaseId) {
							return errors.New("autonomous database was replaced during a password update")
						}
						return nil
					},
				),
			},
			{
				Config:          updatedConfig,
				ConfigVariables: updatedVariables,
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
				Config:          updatedConfig,
				ConfigVariables: updatedVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config:          replacementConfig,
				ConfigVariables: updatedVariables,
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
	displayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, "NONE"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", displayName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
				),
			},
		},
	})
}

func TestAccODBAutonomousDatabase_adminPasswordSource(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var database odbtypes.AutonomousDatabase
	resourceName := "aws_odb_autonomous_database.test"
	dataSourceName := "data.aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabaseAdminPasswordSourcePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutonomousDatabaseConfigAdminPasswordSource(displayName, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &database),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn", os.Getenv(testAccAutonomousDatabaseAdminPasswordSecretARNEnv)),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn", os.Getenv(testAccAutonomousDatabaseAdminPasswordSecretRoleARNEnv)),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type", os.Getenv(testAccAutonomousDatabaseExternalIDTypeEnv)),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSource},
			},
		},
	})
}

func TestAccODBAutonomousDatabase_adminPasswordSourceWithIntegration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var databaseBeforeUpdate, databaseAfterUpdate odbtypes.AutonomousDatabase
	const (
		integrationResourceName = "aws_odb_autonomous_database_secrets_manager_integration.test"
		resourceName            = "aws_odb_autonomous_database.test"
		dataSourceName          = "data.aws_odb_autonomous_database.test"
	)
	displayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	acctest.Test(ctx, t, resource.TestCase{ // nosemgrep:ci.semgrep.acctest.testcase-use-paralleltest -- account-wide integration tests must remain serialized
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabaseSecretsManagerIntegrationPreCheck(ctx, t)
			testAccAutonomousDatabaseAdminPasswordSourcePrerequisitesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAutonomousDatabaseDestroy(ctx, t),
			testAccCheckAutonomousDatabaseSecretsManagerIntegrationDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccAutonomousDatabaseSecretsManagerIntegrationConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(integrationResourceName, names.AttrStatus, string(odbtypes.OciIamRoleStatusAvailable)),
					resource.TestCheckResourceAttrSet(integrationResourceName, names.AttrRoleARN),
				),
			},
			{
				Config: testAccAutonomousDatabaseConfigAdminPasswordSourceWithIntegration(displayName, dbName, "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseBeforeUpdate),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn", os.Getenv(testAccAutonomousDatabaseAdminPasswordSecretARNEnv)),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn", os.Getenv(testAccAutonomousDatabaseAdminPasswordSecretRoleARNEnv)),
					resource.TestCheckResourceAttr(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type", os.Getenv(testAccAutonomousDatabaseExternalIDTypeEnv)),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.secret_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.iam_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type", dataSourceName, "admin_password_source.0.customer_managed_aws_secret.0.external_id_type"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "initial"),
				),
			},
			{
				Config: testAccAutonomousDatabaseConfigAdminPasswordSourceWithIntegration(displayName, dbName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutonomousDatabaseExists(ctx, t, resourceName, &databaseAfterUpdate),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "updated"),
					func(*terraform.State) error {
						if aws.ToString(databaseBeforeUpdate.AutonomousDatabaseId) != aws.ToString(databaseAfterUpdate.AutonomousDatabaseId) {
							return errors.New("Autonomous Database was replaced during a tag update")
						}
						return nil
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSource},
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
	displayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseAdminPasswordEnv)
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseNetworkIDEnv)
	testAccAutonomousDatabaseServicePreCheck(ctx, t)
}

func testAccAutonomousDatabaseAdminPasswordSourcePreCheck(ctx context.Context, t *testing.T) {
	testAccAutonomousDatabaseAdminPasswordSourcePrerequisitesPreCheck(ctx, t)

	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
	role, err := tfodb.FindAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
	if retry.NotFound(err) {
		t.Skip("skipping acceptance testing: AWS Secrets Manager integration is not enabled")
	}
	if err != nil {
		t.Fatalf("unexpected Secrets Manager integration PreCheck error: %s", err)
	}
	if role.Status != odbtypes.OciIamRoleStatusAvailable {
		t.Skipf("skipping acceptance testing: AWS Secrets Manager integration status is %s", role.Status)
	}
}

func testAccAutonomousDatabaseAdminPasswordSourcePrerequisitesPreCheck(ctx context.Context, t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseAdminPasswordSecretARNEnv)
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseAdminPasswordSecretRoleARNEnv)
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseExternalIDTypeEnv)
	acctest.SkipIfEnvVarNotSet(t, testAccAutonomousDatabaseNetworkIDEnv)
	testAccAutonomousDatabaseServicePreCheck(ctx, t)
}

func testAccAutonomousDatabaseServicePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
	input := odb.ListAutonomousDatabasesInput{}
	_, err := conn.ListAutonomousDatabases(ctx, &input)
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

func testAccAutonomousDatabaseAdminPasswordSourceConfigPrerequisites() string {
	return `
variable "odb_test_admin_password_secret_arn" {
  type = string
}

variable "odb_test_admin_password_secret_role_arn" {
  type = string
}

variable "odb_test_external_id_type" {
  type = string
}

variable "odb_test_network_id" {
  type = string
}
`
}

func testAccAutonomousDatabaseConfigBasic(displayName, dbName string, computeCount float64, characterSet, environment string) string {
	return testAccAutonomousDatabaseConfigBasicNamed("test", displayName, dbName, computeCount, characterSet, environment)
}

func testAccAutonomousDatabaseConfigBasicNamed(resourceName, displayName, dbName string, computeCount float64, characterSet, environment string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigPrerequisites(),
		fmt.Sprintf(`
resource "aws_odb_autonomous_database" %[6]q {
  admin_password           = var.odb_test_admin_password
  character_set            = %[1]q
  compute_count            = %[2]g
  data_storage_size_in_tbs = 1
  db_name                  = %[3]q
  db_workload              = "OLTP"
  display_name             = %[4]q
  license_model            = "LICENSE_INCLUDED"
  odb_network_id           = var.odb_test_network_id
  source                   = "NONE"

  tags = {
    Environment = %[5]q
  }
}
`, characterSet, computeCount, dbName, displayName, environment, resourceName),
	)
}

func testAccAutonomousDatabaseConfigImport(basicConfig string) string {
	return acctest.ConfigCompose(basicConfig, `
variable "import_id" {
  type = string
}

removed {
  from = aws_odb_autonomous_database.bootstrap

  lifecycle {
    destroy = false
  }
}

import {
  to = aws_odb_autonomous_database.test
  id = var.import_id
}
`)
}

func testAccAutonomousDatabaseConfigAdminPasswordSource(displayName, dbName string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseAdminPasswordSourceConfigPrerequisites(),
		fmt.Sprintf(`
resource "aws_odb_autonomous_database" "test" {
  character_set            = "AL32UTF8"
  compute_count            = 2
  data_storage_size_in_tbs = 1
  db_name                  = %[1]q
  db_workload              = "OLTP"
  display_name             = %[2]q
  license_model            = "LICENSE_INCLUDED"
  odb_network_id           = var.odb_test_network_id
  source                   = "NONE"

  admin_password_source {
    customer_managed_aws_secret {
      external_id_type = var.odb_test_external_id_type
      iam_role_arn     = var.odb_test_admin_password_secret_role_arn
      secret_arn       = var.odb_test_admin_password_secret_arn
    }
  }
}
`, dbName, displayName),
		`
data "aws_odb_autonomous_database" "test" {
  id = aws_odb_autonomous_database.test.id
}
`,
	)
}

func testAccAutonomousDatabaseConfigAdminPasswordSourceWithIntegration(displayName, dbName, environment string) string {
	return acctest.ConfigCompose(
		testAccAutonomousDatabaseAdminPasswordSourceConfigPrerequisites(),
		testAccAutonomousDatabaseSecretsManagerIntegrationConfigBasic(),
		fmt.Sprintf(`
resource "aws_odb_autonomous_database" "test" {
  character_set            = "AL32UTF8"
  compute_count            = 2
  data_storage_size_in_tbs = 1
  db_name                  = %[1]q
  db_workload              = "OLTP"
  display_name             = %[2]q
  license_model            = "LICENSE_INCLUDED"
  odb_network_id           = var.odb_test_network_id
  source                   = "NONE"

  admin_password_source {
    customer_managed_aws_secret {
      external_id_type = var.odb_test_external_id_type
      iam_role_arn     = var.odb_test_admin_password_secret_role_arn
      secret_arn       = var.odb_test_admin_password_secret_arn
    }
  }

  tags = {
    Environment = %[3]q
  }

  depends_on = [aws_odb_autonomous_database_secrets_manager_integration.test]
}
`, dbName, displayName, environment),
		`
data "aws_odb_autonomous_database" "test" {
  id = aws_odb_autonomous_database.test.id
}
`,
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
