// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSTenantDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_tenant_database.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckTenantDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_basic(rName, "TFPDB1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", regexache.MustCompile(`tenant-database:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "character_set_name"),
					resource.TestCheckResourceAttrPair(resourceName, "db_instance_identifier", "aws_db_instance.test", names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "tfacctest"),
					resource.TestCheckResourceAttrSet(resourceName, "nchar_character_set_name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "TFPDB1"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_database_resource_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "tenant_database_resource_id"),
				ImportStateVerifyIdentifierAttribute: "tenant_database_resource_id",
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_user_password_wo_version",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSTenantDatabase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_tenant_database.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckTenantDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_basic(rName, "TFPDB1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, tfrds.ResourceTenantDatabase, resourceName, tenantDatabaseDisappearsStateFunc),
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

// TestAccRDSTenantDatabase_update covers in-place changes: rename, password rotation,
// switching to a Secrets Manager managed password, and tags.
func TestAccRDSTenantDatabase_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_tenant_database.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckTenantDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_password(rName, "TFPDB1", 1, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "TFPDB1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				// Rename, rotate the password, and change tags in one apply.
				Config: testAccTenantDatabaseConfig_password(rName, "TFPDB2", 2, acctest.CtKey2, acctest.CtValue2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v2),
					testAccCheckTenantDatabaseNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "TFPDB2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTenantDatabaseConfig_manageMasterUserPassword(rName, "TFPDB2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v3),
					testAccCheckTenantDatabaseNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "tenant_database_resource_id"),
				ImportStateVerifyIdentifierAttribute: "tenant_database_resource_id",
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"skip_final_snapshot",
				},
			},
		},
	})
}

// TestAccRDSTenantDatabase_multiple creates tenant databases concurrently on one DB instance,
// which RDS processes one at a time, and covers character sets.
// TestAccRDSTenantDatabase_manageMasterUserPasswordKMSKey verifies that the KMS key can't be changed on a managed
// secret, that it's used again when management is turned off and back on, and that management is adopted after import.
func TestAccRDSTenantDatabase_manageMasterUserPasswordKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4, v5 types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_tenant_database.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckTenantDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "master_user_secret.0.kms_key_id", "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				Config:      testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 1, "tfacctest"),
				ExpectError: regexache.MustCompile(`Unsupported change`),
			},
			{
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, false, 0, "tfacctest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v2),
					testAccCheckTenantDatabaseNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", "0"),
				),
			},
			{
				// Turning management back on uses the configured key even though it didn't change.
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v3),
					testAccCheckTenantDatabaseNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttrPair(resourceName, "master_user_secret.0.kms_key_id", "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				// Forget the tenant database so it can be imported again.
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKeyRemoved(rName),
			},
			{
				// Imported state has no key, and manage_master_user_password is inferred from the secret.
				Config:             testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateIdFunc:  func(*terraform.State) (string, error) { return aws.ToString(v3.TenantDatabaseResourceId), nil },
				ImportStatePersist: true,
			},
			{
				// After import, only the secret's key ARN is accepted.
				Config:      testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 1, "tfacctest"),
				ExpectError: regexache.MustCompile(`Unsupported change`),
			},
			{
				// After import, the matching key ARN is recorded without resending it.
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v4),
					testAccCheckTenantDatabaseNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "master_user_secret.0.kms_key_id", "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKeyRemoved(rName),
			},
			{
				Config:             testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateIdFunc:  func(*terraform.State) (string, error) { return aws.ToString(v4.TenantDatabaseResourceId), nil },
				ImportStatePersist: true,
			},
			{
				// After import, a password with the flag omitted turns management off.
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKeyPassword(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v5),
					testAccCheckTenantDatabaseNotRecreated(&v4, &v5),
					resource.TestCheckNoResourceAttr(resourceName, "manage_master_user_password"),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", "0"),
				),
			},
			{
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 0, "tfacctest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "master_user_secret.0.kms_key_id", "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				// A key change is allowed when the tenant database is replaced.
				Config: testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName, true, 1, "tfacctest2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "master_username", "tfacctest2"),
					resource.TestCheckResourceAttrPair(resourceName, "master_user_secret.0.kms_key_id", "aws_kms_key.test.1", names.AttrARN),
				),
			},
		},
	})
}

// TestAccRDSTenantDatabase_passwordWithManagedPassword verifies that a password is rejected when management is on.
// It fails during validation, so no infrastructure is created.
func TestAccRDSTenantDatabase_passwordWithManagedPassword(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccTenantDatabaseConfig_passwordWithManagedPassword(rName),
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccRDSTenantDatabase_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_rds_tenant_database.test1"
	resource2Name := "aws_rds_tenant_database.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckTenantDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resource1Name, &v1),
					testAccCheckTenantDatabaseExists(ctx, t, resource2Name, &v2),
					resource.TestCheckResourceAttr(resource1Name, "character_set_name", "WE8ISO8859P15"),
					resource.TestCheckResourceAttr(resource2Name, "nchar_character_set_name", "UTF8"),
				),
			},
		},
	})
}

// TestAccRDSTenantDatabase_finalSnapshot verifies that deleting a tenant database takes a final snapshot by default.
func TestAccRDSTenantDatabase_finalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_tenant_database.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckTenantDatabaseDestroy(ctx, t),
			testAccCheckTenantDatabaseFinalSnapshot(ctx, t, rName),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDatabaseConfig_finalSnapshot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("skip_final_snapshot"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccCheckTenantDatabaseExists(ctx context.Context, t *testing.T, n string, v *types.TenantDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindTenantDatabaseByID(ctx, conn, rs.Primary.Attributes["tenant_database_resource_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTenantDatabaseDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_tenant_database" {
				continue
			}

			_, err := tfrds.FindTenantDatabaseByID(ctx, conn, rs.Primary.Attributes["tenant_database_resource_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Tenant Database %s still exists", rs.Primary.Attributes["tenant_database_resource_id"])
		}

		return nil
	}
}

// tenantDatabaseDisappearsStateFunc also sets skip_final_snapshot, which the default state function (string attributes only) omits.
func tenantDatabaseDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	for _, name := range []string{"db_instance_identifier", "tenant_database_resource_id", "tenant_db_name"} {
		if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(name), is.Attributes[name])); err != nil {
			return err
		}
	}

	return fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("skip_final_snapshot"), true))
}

func testAccCheckTenantDatabaseNotRecreated(before, after *types.TenantDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.TenantDatabaseResourceId), aws.ToString(after.TenantDatabaseResourceId); before != after {
			return fmt.Errorf("RDS Tenant Database (%s/%s) recreated", before, after)
		}

		return nil
	}
}

// testAccCheckTenantDatabaseFinalSnapshot checks that the final snapshot exists, then deletes it.
func testAccCheckTenantDatabaseFinalSnapshot(ctx context.Context, t *testing.T, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		if _, err := tfrds.FindDBSnapshotByID(ctx, conn, rName); err != nil {
			return fmt.Errorf("reading RDS Tenant Database final snapshot (%s): %w", rName, err)
		}

		input := rds.DeleteDBSnapshotInput{
			DBSnapshotIdentifier: aws.String(rName),
		}
		if _, err := conn.DeleteDBSnapshot(ctx, &input); err != nil {
			return fmt.Errorf("deleting RDS Tenant Database final snapshot (%s): %w", rName, err)
		}

		return nil
	}
}

// testAccTenantDatabaseConfig_base creates an RDS for Oracle CDB instance in the multi-tenant configuration.
// The instance's initial tenant database is named TFTEST.
func testAccTenantDatabaseConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRandomPassword(),
		fmt.Sprintf(`
# The multi-tenant configuration requires Oracle Database 19c RU 2022-01 or higher.
data "aws_rds_engine_version" "test" {
  engine  = %[1]q
  version = "19"
  latest  = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.test.engine
  engine_version = data.aws_rds_engine_version.test.version_actual
  license_model  = "bring-your-own-license"
  storage_type   = "gp3"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage   = 20
  apply_immediately   = true
  db_name             = "TFTEST"
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = %[3]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  multi_tenant        = true
  password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version = 1
  skip_final_snapshot = true
  storage_type        = data.aws_rds_orderable_db_instance.test.storage_type
  username            = "tfacctest"
}
`, tfrds.InstanceEngineOracleStandard2CDB, mainInstanceClasses, rName))
}

func testAccTenantDatabaseConfig_basic(rName, tenantDBName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  tenant_db_name                  = %[1]q
}
`, tenantDBName))
}

func testAccTenantDatabaseConfig_password(rName, tenantDBName string, passwordVersion int, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = %[2]d
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  tenant_db_name                  = %[1]q

  tags = {
    %[3]q = %[4]q
  }
}
`, tenantDBName, passwordVersion, tagKey, tagValue))
}

func testAccTenantDatabaseConfig_manageMasterUserPassword(rName, tenantDBName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier      = aws_db_instance.test.identifier
  manage_master_user_password = true
  master_username             = "tfacctest"
  skip_final_snapshot         = true
  tenant_db_name              = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, tenantDBName, acctest.CtKey2, acctest.CtValue2))
}

func testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKey(rName string, managed bool, keyIndex int, masterUsername string) string {
	// Turning off management requires a password. Setting the flag explicitly also covers false with a password.
	password := fmt.Sprintf(`
  manage_master_user_password = %[1]t`, managed)
	if !managed {
		password += `
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1`
	}

	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_rds_tenant_database" "test" {
  db_instance_identifier        = aws_db_instance.test.identifier
  master_user_secret_kms_key_id = aws_kms_key.test[%[2]d].arn
  master_username               = %[4]q
  skip_final_snapshot           = true
  tenant_db_name                = "TFPDB1"
%[3]s
}
`, rName, keyIndex, password, masterUsername))
}

func testAccTenantDatabaseConfig_passwordWithManagedPassword(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier      = %[1]q
  manage_master_user_password = %[2]t
  master_user_password        = "avoid-plaintext-passwords"
  master_username             = "tfacctest"
  skip_final_snapshot         = %[2]t
  tenant_db_name              = "TFPDB1"
}
`, rName, true)
}

func testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKeyRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

removed {
  from = aws_rds_tenant_database.test

  lifecycle {
    destroy = %[2]t
  }
}
`, rName, false))
}

func testAccTenantDatabaseConfig_manageMasterUserPasswordKMSKeyPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_rds_tenant_database" "test" {
  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  tenant_db_name                  = "TFPDB1"
}
`, rName))
}

func testAccTenantDatabaseConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		`
resource "aws_rds_tenant_database" "test1" {
  character_set_name              = "WE8ISO8859P15"
  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest1"
  skip_final_snapshot             = true
  tenant_db_name                  = "TFPDB1"
}

resource "aws_rds_tenant_database" "test2" {
  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest2"
  nchar_character_set_name        = "UTF8"
  skip_final_snapshot             = true
  tenant_db_name                  = "TFPDB2"
}
`)
}

func testAccTenantDatabaseConfig_finalSnapshot(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier          = aws_db_instance.test.identifier
  final_snapshot_identifier       = %[1]q
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest"
  tenant_db_name                  = "TFPDB1"
}
`, rName))
}
