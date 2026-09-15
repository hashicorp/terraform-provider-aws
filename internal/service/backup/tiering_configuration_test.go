// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupTieringConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TieringConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_backup_tiering_configuration.test"
	vaultResourceName := "aws_backup_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTieringConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "backup_vault_name", vaultResourceName, names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("backup", regexache.MustCompile(`tiering-configuration:`+name+`-.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrLastUpdatedTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_selection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrResources:             knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("*")}),
							names.AttrResourceType:          knownvalue.StringExact("S3"),
							"tiering_down_settings_in_days": knownvalue.Int32Exact(60),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupTieringConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TieringConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_backup_tiering_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTieringConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbackup.ResourceTieringConfiguration, resourceName),
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

func TestAccBackupTieringConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TieringConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_backup_tiering_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTieringConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.resources.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_selection.0.resources.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tiering_down_settings_in_days", "60"),
				),
			},
			{
				Config: testAccTieringConfigurationConfig_updated(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.resources.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "resource_selection.0.resources.*", bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tiering_down_settings_in_days", "90"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccTieringConfigurationConfig_allVaults(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.resources.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_selection.0.resources.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tiering_down_settings_in_days", "120"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupTieringConfiguration_multipleResourceSelections(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TieringConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_backup_tiering_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTieringConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_multipleResourceSelections(rName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.resource_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.resources.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "resource_selection.0.resources.*", "aws_s3_bucket.test.0", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tiering_down_settings_in_days", "60"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.1.resource_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.1.resources.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "resource_selection.1.resources.*", "aws_s3_bucket.test.1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.1.tiering_down_settings_in_days", "120"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupTieringConfiguration_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name := strings.ReplaceAll(rName, "-", "_")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTieringConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccTieringConfigurationConfig_missingResourceSelection(rName, name),
				ExpectError: regexache.MustCompile(`resource_selection.*must have a configuration value`),
			},
			{
				Config:      testAccTieringConfigurationConfig_invalidResourceType(rName, name),
				ExpectError: regexache.MustCompile(`resource_type value must be one of: \["S3"\]`),
			},
			{
				Config:      testAccTieringConfigurationConfig_tooManyResources(rName, name),
				ExpectError: regexache.MustCompile(`must select at most 100 resources in total, got 102`),
			},
			{
				Config:      testAccTieringConfigurationConfig_allVaultsSpecificResource(rName, name),
				ExpectError: regexache.MustCompile(`resources must select all resources when backup_vault_name is "\*"`),
			},
		},
	})
}

func testAccCheckTieringConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_tiering_configuration" {
				continue
			}

			_, err := tfbackup.FindTieringConfigurationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Tiering Configuration %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckTieringConfigurationExists(ctx context.Context, t *testing.T, n string, v *awstypes.TieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		output, err := tfbackup.FindTieringConfigurationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckTieringConfiguration(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)

	conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

	input := backup.ListTieringConfigurationsInput{}
	_, err := conn.ListTieringConfigurations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTieringConfigurationConfig_basic(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_missingResourceSelection(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name
}
`, rName, name)
}

func testAccTieringConfigurationConfig_invalidResourceType(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "EBS"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_tooManyResources(rName, name string) string {
	// The ARNs are literals so that they are known when the resource configuration is validated.
	//lintignore:AWSAT005
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = formatlist("arn:aws:s3:::%[1]s-a-%%03d", range(51))
    tiering_down_settings_in_days = 60
  }

  resource_selection {
    resource_type                 = "S3"
    resources                     = formatlist("arn:aws:s3:::%[1]s-b-%%03d", range(51))
    tiering_down_settings_in_days = 90
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_updated(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = [aws_s3_bucket.test.arn]
    tiering_down_settings_in_days = 90
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_allVaults(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = "*"

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 120
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_allVaultsSpecificResource(rName, name string) string {
	//lintignore:AWSAT005
	return fmt.Sprintf(`
resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = "*"

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["arn:aws:s3:::%[1]s"]
    tiering_down_settings_in_days = 60
  }
}
`, rName, name)
}

func testAccTieringConfigurationConfig_multipleResourceSelections(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  count = 2

  bucket = "%[1]s-${count.index}"
}

resource "aws_backup_tiering_configuration" "test" {
  name              = %[2]q
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = [aws_s3_bucket.test[0].arn]
    tiering_down_settings_in_days = 60
  }

  resource_selection {
    resource_type                 = "S3"
    resources                     = [aws_s3_bucket.test[1].arn]
    tiering_down_settings_in_days = 120
  }
}
`, rName, name)
}
