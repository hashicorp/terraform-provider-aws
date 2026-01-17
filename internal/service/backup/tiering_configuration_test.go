// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupTieringConfiguration_backupVaultName(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, vaultName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", vaultName1),
				),
			},
			{
				Config: testAccTieringConfigurationConfig_basic(rName, vaultName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", vaultName2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBackupTieringConfiguration_backupVaultName_wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_wildcardVault(rName, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "tiering_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_selection.*", map[string]string{
						names.AttrResourceType:          "S3",
						"tiering_down_settings_in_days": "180",
					}),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
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

func TestAccBackupTieringConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "backup_vault_name", vaultName),
					resource.TestCheckResourceAttr(resourceName, "tiering_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_selection.*", map[string]string{
						names.AttrResourceType:          "S3",
						"tiering_down_settings_in_days": "90",
					}),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
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

func TestAccBackupTieringConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbackup.ResourceTieringConfiguration, resourceName),
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

func TestAccBackupTieringConfiguration_name(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName1 := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	rName2 := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName1, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "tiering_configuration_name", rName1),
				),
			},
			{
				Config: testAccTieringConfigurationConfig_basic(rName2, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "tiering_configuration_name", rName2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccBackupTieringConfiguration_resourceSelection_multipleResources(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBucketName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_multipleResourceSelection(rName, vaultName, rBucketName, rBucketName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "2"),
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

func TestAccBackupTieringConfiguration_resourceSelection(t *testing.T) {
	ctx := acctest.Context(t)
	var tieringConfiguration awstypes.TieringConfiguration
	resourceName := "aws_backup_tiering_configuration.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))
	vaultName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTieringConfigurationConfig_basic(rName, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_selection.*", map[string]string{
						"tiering_down_settings_in_days": "90",
					}),
				),
			},
			{
				Config: testAccTieringConfigurationConfig_updated(rName, vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTieringConfigurationExists(ctx, resourceName, &tieringConfiguration),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_selection.*", map[string]string{
						"tiering_down_settings_in_days": "120",
					}),
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

func testAccCheckTieringConfigurationExists(ctx context.Context, n string, v *awstypes.TieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Tiering Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfresource.RetryWhenNotFound(ctx, 1*time.Minute, func(ctx context.Context) (any, error) {
			return tfbackup.FindTieringConfigurationByName(ctx, conn, rs.Primary.ID)
		})
		if err != nil {
			return fmt.Errorf("error finding Tiering Configuration %q: %w", rs.Primary.ID, err)
		}

		*v = *(output.(*awstypes.TieringConfiguration))

		return nil
	}
}

func testAccCheckTieringConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_tiering_configuration" {
				continue
			}

			_, err := tfbackup.FindTieringConfigurationByName(ctx, conn, rs.Primary.ID)

			if errs.IsA[*sdkretry.NotFoundError](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Tiering Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTieringConfigurationConfig_basic(rName, vaultName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[2]q
}

resource "aws_backup_tiering_configuration" "test" {
  tiering_configuration_name = %[1]q
  backup_vault_name          = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 90
  }
}
`, rName, vaultName)
}

func testAccTieringConfigurationConfig_updated(rName, vaultName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[2]q
}

resource "aws_backup_tiering_configuration" "test" {
  tiering_configuration_name = %[1]q
  backup_vault_name          = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 120
  }
}
`, rName, vaultName)
}

func testAccTieringConfigurationConfig_multipleResourceSelection(rName, vaultName, rBucketName, rBucketName2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}

resource "aws_s3_bucket" "test2" {
  bucket = %[4]q
}

resource "aws_backup_vault" "test" {
  name = %[2]q
}

resource "aws_backup_tiering_configuration" "test" {
  tiering_configuration_name = %[1]q
  backup_vault_name          = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = [aws_s3_bucket.test.arn]
    tiering_down_settings_in_days = 90
  }

  resource_selection {
    resource_type                 = "S3"
    resources                     = [aws_s3_bucket.test2.arn]
    tiering_down_settings_in_days = 60
  }

  depends_on = [aws_s3_bucket.test, aws_s3_bucket.test2]
}
`, rName, vaultName, rBucketName, rBucketName2)
}

func testAccTieringConfigurationConfig_wildcardVault(rName, vaultName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[2]q
}

resource "aws_backup_tiering_configuration" "test" {
  tiering_configuration_name = %[1]q
  backup_vault_name          = "*"

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 180
  }
}
`, rName, vaultName)
}
