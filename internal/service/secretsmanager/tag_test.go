// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(acctest.CtKey1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact(acctest.CtValue1)),
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

func TestAccSecretsManagerTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecretsmanager.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSecretsManagerTag_Value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_tag.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(acctest.CtKey1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact(acctest.CtValue1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig_basic(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(acctest.CtKey1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact(acctest.CtValue1Updated)),
				},
			},
		},
	})
}

func TestAccSecretsManagerTag_managedSecrets(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_secretsmanager_tag.rds"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_managedSecrets(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(acctest.CtKey1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact(acctest.CtValue1)),
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

func testAccTagConfig_basic(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_secretsmanager_tag" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  key       = %[2]q
  value     = %[3]q
}
`, rName, key, value)
}

func testAccTagConfig_managedSecrets(key string, value string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "example" {
  identifier = "exampledb-rds-instance"

  engine         = "mysql"
  engine_version = "8.0.43"

  instance_class = "db.t3.micro"

  allocated_storage     = 20
  max_allocated_storage = 0
  storage_type          = "gp2"
  storage_encrypted     = false

  db_name                     = "exampledb"
  username                    = "dbadmin"
  manage_master_user_password = true
  port                        = 3306

  publicly_accessible = false

  multi_az = false

  backup_retention_period = 0
  skip_final_snapshot     = true
  deletion_protection     = false

  monitoring_interval          = 0
  performance_insights_enabled = false

  apply_immediately          = true
  auto_minor_version_upgrade = true
  maintenance_window         = "mon:00:00-mon:03:00"

  tags = {
    Name = "exampledb-rds"
  }
}

data "aws_secretsmanager_secret" "rds" {
  arn = aws_db_instance.example.master_user_secret[0].secret_arn
}

resource "aws_secretsmanager_tag" "rds" {
  secret_id = data.aws_secretsmanager_secret.rds.id
  key       = %[1]q
  value     = %[2]q
}
`, key, value)
}
