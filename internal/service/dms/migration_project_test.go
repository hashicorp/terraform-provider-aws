// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const resNameMigrationProject = "Migration Project"

func TestAccDMSMigrationProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_migration_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckMigrationProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMigrationProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMigrationProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMigrationProjectExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dms", regexache.MustCompile(`migration-project:.+$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrCreationTime, regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$`)),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_arn", "aws_dms_instance_profile.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "instance_profile_name"),
					resource.TestCheckResourceAttr(resourceName, "source_data_provider_descriptor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "source_data_provider_descriptor.0.data_provider_arn", "aws_dms_data_provider.source", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target_data_provider_descriptor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_data_provider_descriptor.0.data_provider_arn", "aws_dms_data_provider.target", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDMSMigrationProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_migration_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckMigrationProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMigrationProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMigrationProjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMigrationProjectExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceMigrationProject, resourceName),
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

func TestAccDMSMigrationProject_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_migration_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DMS)
			testAccPreCheckMigrationProject(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMigrationProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMigrationProjectConfig_full(rName, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMigrationProjectExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first"),
				),
			},
			{
				Config: testAccMigrationProjectConfig_full(rName, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMigrationProjectExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckMigrationProjectDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_migration_project" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
			arn := rs.Primary.Attributes[names.AttrARN]
			_, err := tfdms.FindMigrationProjectByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DMS, create.ErrActionCheckingDestroyed, resNameMigrationProject, arn, err)
			}

			return create.Error(names.DMS, create.ErrActionCheckingDestroyed, resNameMigrationProject, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMigrationProjectExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameMigrationProject, name, errors.New("not found"))
		}

		arn := rs.Primary.Attributes[names.AttrARN]
		if arn == "" {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameMigrationProject, name, errors.New("arn not set"))
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
		_, err := tfdms.FindMigrationProjectByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, resNameMigrationProject, arn, err)
		}

		return nil
	}
}

func testAccPreCheckMigrationProject(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
	var input databasemigrationservice.DescribeMigrationProjectsInput
	_, err := conn.DescribeMigrationProjects(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMigrationProjectConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_dms_instance_profile" "test" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "secretsmanager:*"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_secretsmanager_secret" "source" {
  name                    = "%[1]s-source"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "source" {
  secret_id     = aws_secretsmanager_secret.source.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
}

resource "aws_secretsmanager_secret" "target" {
  name                    = "%[1]s-target"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "target" {
  secret_id     = aws_secretsmanager_secret.target.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
}

resource "aws_dms_data_provider" "source" {
  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "%[1]s-source.example.com"
      ssl_mode      = "none"
    }
  }
}

resource "aws_dms_data_provider" "target" {
  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "%[1]s-target.example.com"
      ssl_mode      = "none"
    }
  }
}
`, rName)
}

func testAccMigrationProjectConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMigrationProjectConfig_base(rName), `
resource "aws_dms_migration_project" "test" {
  instance_profile_arn = aws_dms_instance_profile.test.arn

  source_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.source.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.target.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.target.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`)
}

func testAccMigrationProjectConfig_full(rName, description string) string {
	return acctest.ConfigCompose(testAccMigrationProjectConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_migration_project" "test" {
  name                 = %[1]q
  description          = %[2]q
  instance_profile_arn = aws_dms_instance_profile.test.arn

  source_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.source.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.target.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.target.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, description))
}
