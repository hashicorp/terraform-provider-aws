// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
				Config: testAccTenantDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_instance_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "TESTPDB"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tfacctest"),
					resource.TestCheckResourceAttr(resourceName, "character_set_name", "AL32UTF8"),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "AL16UTF16"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_database_resource_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "available"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
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
				Config: testAccTenantDatabaseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceTenantDatabase(), resourceName),
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

func TestAccRDSTenantDatabase_tenantDBName(t *testing.T) {
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
				Config: testAccTenantDatabaseConfig_tenantDBName(rName, "MYPDB1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "MYPDB1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
			{
				Config: testAccTenantDatabaseConfig_tenantDBName(rName, "MYPDB2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tenant_db_name", "MYPDB2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
		},
	})
}

func TestAccRDSTenantDatabase_masterPassword(t *testing.T) {
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
				Config: testAccTenantDatabaseConfig_masterPassword(rName, "tFaccPass1!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
			{
				Config: testAccTenantDatabaseConfig_masterPassword(rName, "tFaccPass2!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func TestAccRDSTenantDatabase_manageMasterUserPassword(t *testing.T) {
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
				Config: testAccTenantDatabaseConfig_manageMasterUserPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.0.secret_status", "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
		},
	})
}

func TestAccRDSTenantDatabase_characterSet(t *testing.T) {
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
				Config: testAccTenantDatabaseConfig_characterSet(rName, "UTF8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "character_set_name", "UTF8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
		},
	})
}

func TestAccRDSTenantDatabase_ncharCharacterSet(t *testing.T) {
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
				Config: testAccTenantDatabaseConfig_ncharCharacterSet(rName, "UTF8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantDatabaseExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "UTF8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password",
				},
			},
		},
	})
}

// TestAccRDSTenantDatabase_multipleParallel verifies that two PDBs on the
// same CDB can be destroyed concurrently without the second delete failing
// because the CDB is in a non-available state while the first delete completes.
// This exercises the RetryWhenIsA fix in resourceTenantDatabaseDelete.
func TestAccRDSTenantDatabase_multipleParallel(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pdb1, pdb2 types.TenantDatabase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_rds_tenant_database.pdb1"
	resource2Name := "aws_rds_tenant_database.pdb2"

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
					testAccCheckTenantDatabaseExists(ctx, t, resource1Name, &pdb1),
					testAccCheckTenantDatabaseExists(ctx, t, resource2Name, &pdb2),
					resource.TestCheckResourceAttr(resource1Name, "tenant_db_name", "MYPDB1"),
					resource.TestCheckResourceAttr(resource2Name, "tenant_db_name", "MYPDB2"),
					resource.TestCheckResourceAttr(resource1Name, "db_instance_identifier", rName),
					resource.TestCheckResourceAttr(resource2Name, "db_instance_identifier", rName),
				),
			},
			// No explicit destroy step: the framework destroys all resources at
			// end of test. Absence of an error confirms the parallel-delete retry works.
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

		output, err := tfrds.FindTenantDatabaseByID(ctx, conn, rs.Primary.ID)
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

			_, err := tfrds.FindTenantDatabaseByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Tenant Database %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccTenantDatabaseConfig_base provisions the CDB instance shared by all
// tenant database configs. Oracle CDB requires a subnet group, a CDB parameter
// group family, and multi_tenant = true on the db_instance.
func testAccTenantDatabaseConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRandomPassword(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = %[1]q
  license_model = "bring-your-own-license"
  storage_type  = "gp3"

  preferred_instance_classes = [%[2]s]

  # CDB multi_tenant requires 19c RU 2021-10 or later; exclude older RUs.
  preferred_engine_versions = [
    "19.0.0.0.ru-2024-10.rur-2024-10.r1",
    "19.0.0.0.ru-2024-07.rur-2024-07.r1",
    "19.0.0.0.ru-2024-04.rur-2024-04.r1",
    "19.0.0.0.ru-2024-01.rur-2024-01.r1",
    "19.0.0.0.ru-2023-10.rur-2023-10.r1",
    "19.0.0.0.ru-2023-07.rur-2023-07.r1",
    "19.0.0.0.ru-2023-04.rur-2023-04.r1",
    "19.0.0.0.ru-2023-01.rur-2023-01.r1",
    "19.0.0.0.ru-2022-10.rur-2022-10.r1",
    "19.0.0.0.ru-2022-07.rur-2022-07.r1",
    "19.0.0.0.ru-2022-04.rur-2022-04.r1",
    "19.0.0.0.ru-2022-01.rur-2022-01.r1",
    "19.0.0.0.ru-2021-10.rur-2021-10.r1",
  ]
}

resource "aws_db_parameter_group" "test" {
  name   = %[3]q
  family = "oracle-se2-cdb-19"
}

resource "aws_db_instance" "test" {
  identifier           = %[3]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  license_model        = "bring-your-own-license"
  multi_tenant         = true
  allocated_storage    = 200
  storage_type         = "gp3"
  db_subnet_group_name = aws_db_subnet_group.test.name
  parameter_group_name = aws_db_parameter_group.test.name
  password_wo          = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version  = 1
  username             = "tfacctest"
  skip_final_snapshot  = true
}
`, tfrds.InstanceEngineOracleStandard2CDB, mainInstanceClasses, rName))
}

func testAccTenantDatabaseConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = "TESTPDB"
  username               = "tfacctest"
  master_password        = "tFaccPass1!"
}
`)
}

func testAccTenantDatabaseConfig_tenantDBName(rName, pdbName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = %[1]q
  username               = "tfacctest"
  master_password        = "tFaccPass1!"
}
`, pdbName))
}

func testAccTenantDatabaseConfig_masterPassword(rName, password string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = "TESTPDB"
  username               = "tfacctest"
  master_password        = %[1]q
}
`, password))
}

func testAccTenantDatabaseConfig_manageMasterUserPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier      = aws_db_instance.test.identifier
  tenant_db_name              = "TESTPDB"
  username                    = "tfacctest"
  manage_master_user_password = true
}
`)
}

func testAccTenantDatabaseConfig_characterSet(rName, characterSet string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = "TESTPDB"
  username               = "tfacctest"
  master_password        = "tFaccPass1!"
  character_set_name     = %[1]q
}
`, characterSet))
}

func testAccTenantDatabaseConfig_ncharCharacterSet(rName, ncharCharacterSet string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_tenant_database" "test" {
  db_instance_identifier   = aws_db_instance.test.identifier
  tenant_db_name           = "TESTPDB"
  username                 = "tfacctest"
  master_password          = "tFaccPass1!"
  nchar_character_set_name = %[1]q
}
`, ncharCharacterSet))
}

func testAccTenantDatabaseConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccTenantDatabaseConfig_base(rName),
		`
resource "aws_rds_tenant_database" "pdb1" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = "MYPDB1"
  username               = "pdb1admin"
  master_password        = "tFaccPass1!"
}

resource "aws_rds_tenant_database" "pdb2" {
  db_instance_identifier = aws_db_instance.test.identifier
  tenant_db_name         = "MYPDB2"
  username               = "pdb2admin"
  master_password        = "tFaccPass2!"
}
`)
}
