// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "backup", regexache.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                    rName,
						"target_vault_name":            rName,
						names.AttrSchedule:             "cron(0 12 * * ? *)",
						"schedule_expression_timezone": "Etc/UTC",
						"lifecycle.#":                  "0",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
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

func TestAccBackupPlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbackup.ResourcePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupPlan_withRules(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))
	rule1Name := fmt.Sprintf("%s_1", rName)
	rule2Name := fmt.Sprintf("%s_2", rName)
	rule3Name := fmt.Sprintf("%s_3", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_twoRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule1Name,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule2Name,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 6 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_threeRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule1Name,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 6 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule2Name,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule3Name,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 18 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rName,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccBackupPlan_withLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_lifecycleColdStorageAfterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "7",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_lifecycleDeleteAfterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"lifecycle.#":              "1",
						"lifecycle.0.delete_after": "120",
					}),
				),
			},
			{
				Config: testAccPlanConfig_lifecycleColdStorageAfterAndDeleteAfter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
					}),
				),
			},
			{
				Config: testAccPlanConfig_optInToArchiveForSupportedResources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"lifecycle.0.opt_in_to_archive_for_supported_resources": acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":   rName,
						"lifecycle.#": "0",
					}),
				),
			},
		},
	})
}

func TestAccBackupPlan_withRecoveryPointTags(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_recoveryPointTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						names.AttrSchedule:         "cron(0 12 * * ? *)",
						"lifecycle.#":              "0",
						"recovery_point_tags.%":    "3",
						"recovery_point_tags.Name": rName,
						"recovery_point_tags.Key1": "Value1",
						"recovery_point_tags.Key2": "Value2a",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_recoveryPointTagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						names.AttrSchedule:         "cron(0 12 * * ? *)",
						"lifecycle.#":              "0",
						"recovery_point_tags.%":    "3",
						"recovery_point_tags.Name": rName,
						"recovery_point_tags.Key2": "Value2b",
						"recovery_point_tags.Key3": "Value3",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rName,
						"target_vault_name": rName,
						names.AttrSchedule:  "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccBackupPlan_RuleCopyAction_sameRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_ruleCopyAction(rName, 30, 180),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"copy_action.#":                  "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_ruleCopyAction(rName, 60, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"copy_action.#":                  "1",
					}),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":     rName,
						"lifecycle.#":   "0",
						"copy_action.#": "0",
					}),
				),
			},
		},
	})
}

func TestAccBackupPlan_RuleCopyAction_noLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_ruleCopyActionNoLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":     rName,
						"lifecycle.#":   "0",
						"copy_action.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_ruleCopyAction(rName, 60, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"copy_action.#":                  "1",
					}),
				),
			},
			{
				Config: testAccPlanConfig_ruleCopyActionNoLifecycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":     rName,
						"lifecycle.#":   "0",
						"copy_action.#": "1",
					}),
				),
			},
		},
	})
}

func TestAccBackupPlan_RuleCopyAction_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_ruleCopyActionMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"copy_action.#":                  "2",
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

func TestAccBackupPlan_RuleCopyAction_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_ruleCopyActionCrossRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
						"copy_action.#":                  "1",
					}),
				),
			},
			{
				Config:            testAccPlanConfig_ruleCopyActionCrossRegion(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupPlan_advancedBackupSetting(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_advancedSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "advanced_backup_setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_backup_setting.*", map[string]string{
						"backup_options.%":          "1",
						"backup_options.WindowsVSS": names.AttrEnabled,
						names.AttrResourceType:      "EC2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_advancedSettingUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "advanced_backup_setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_backup_setting.*", map[string]string{
						"backup_options.%":          "1",
						"backup_options.WindowsVSS": "disabled",
						names.AttrResourceType:      "EC2",
					}),
				),
			},
		},
	})
}

func TestAccBackupPlan_enableContinuousBackup(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_enableContinuous(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "backup", regexache.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						names.AttrSchedule:         "cron(0 12 * * ? *)",
						"enable_continuous_backup": acctest.CtTrue,
						"lifecycle.#":              "1",
						"lifecycle.0.delete_after": "35",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
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

func TestAccBackupPlan_upgradeScheduleExpressionTimezone(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BackupServiceID),
		CheckDestroy: testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.70.0",
					},
				},
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
				),
			},
		},
	})
}

func TestAccBackupPlan_scheduleExpressionTimezone(t *testing.T) {
	ctx := acctest.Context(t)
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_scheduleExpressionTimezone(rName, "Pacific/Tahiti"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"lifecycle.#":                  "0",
						"rule_name":                    rName,
						names.AttrSchedule:             "cron(0 12 * * ? *)",
						"schedule_expression_timezone": "Pacific/Tahiti",
						"target_vault_name":            rName,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_scheduleExpressionTimezone(rName, "Africa/Abidjan"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"lifecycle.#":                  "0",
						"rule_name":                    rName,
						names.AttrSchedule:             "cron(0 12 * * ? *)",
						"schedule_expression_timezone": "Africa/Abidjan",
						"target_vault_name":            rName,
					}),
				),
			},
		},
	})
}

func testAccCheckPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_plan" {
				continue
			}

			_, err := tfbackup.FindPlanByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Plan %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPlanExists(ctx context.Context, n string, v *backup.GetBackupPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindPlanByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPlanConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
}
`, rName)
}

func testAccPlanConfig_optInToArchiveForSupportedResources(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 10 ? * 6L *)"
    lifecycle {
      cold_storage_after                        = 30
      delete_after                              = 180
      opt_in_to_archive_for_supported_resources = true
    }
  }
}
`, rName)
}

func testAccPlanConfig_twoRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = "%[1]s_1"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_2"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 6 * * ? *)"
  }
}
`, rName)
}

func testAccPlanConfig_threeRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = "%[1]s_1"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 6 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_2"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_3"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 18 * * ? *)"
  }
}
`, rName)
}

func testAccPlanConfig_lifecycleColdStorageAfterOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 7
    }
  }
}
`, rName)
}

func testAccPlanConfig_lifecycleDeleteAfterOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      delete_after = 120
    }
  }
}
`, rName)
}

func testAccPlanConfig_lifecycleColdStorageAfterAndDeleteAfter(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }
  }
}
`, rName)
}

func testAccPlanConfig_recoveryPointTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    recovery_point_tags = {
      Name = %[1]q
      Key1 = "Value1"
      Key2 = "Value2a"
    }
  }
}
`, rName)
}

func testAccPlanConfig_recoveryPointTagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    recovery_point_tags = {
      Name = %[1]q
      Key2 = "Value2b"
      Key3 = "Value3"
    }
  }
}
`, rName)
}

func testAccPlanConfig_ruleCopyAction(rName string, coldStorageAfter, deleteAfter int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%[1]s-1"
}

resource "aws_backup_vault" "test2" {
  name = "%[1]s-2"
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }

    copy_action {
      lifecycle {
        cold_storage_after = %[2]d
        delete_after       = %[3]d
      }

      destination_vault_arn = aws_backup_vault.test2.arn
    }
  }
}
`, rName, coldStorageAfter, deleteAfter)
}

func testAccPlanConfig_ruleCopyActionMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%[1]s-1"
}

resource "aws_backup_vault" "test2" {
  name = "%[1]s-2"
}

resource "aws_backup_vault" "test3" {
  name = "%[1]s-3"
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }

    copy_action {
      lifecycle {
        cold_storage_after = 30
        delete_after       = 180
      }

      destination_vault_arn = aws_backup_vault.test2.arn
    }

    copy_action {
      lifecycle {
        cold_storage_after = 60
        delete_after       = 365
      }

      destination_vault_arn = aws_backup_vault.test3.arn
    }
  }
}
`, rName)
}

func testAccPlanConfig_ruleCopyActionCrossRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%[1]s-1"
}

resource "aws_backup_vault" "test2" {
  provider = "awsalternate"
  name     = "%[1]s-2"
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }

    copy_action {
      lifecycle {
        cold_storage_after = 30
        delete_after       = 180
      }

      destination_vault_arn = aws_backup_vault.test2.arn
    }
  }
}
`, rName))
}

func testAccPlanConfig_ruleCopyActionNoLifecycle(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%[1]s-1"
}

resource "aws_backup_vault" "test2" {
  name = "%[1]s-2"
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    copy_action {
      destination_vault_arn = aws_backup_vault.test2.arn
    }
  }
}
`, rName)
}

func testAccPlanConfig_advancedSetting(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }
  }

  advanced_backup_setting {
    backup_options = {
      WindowsVSS = "enabled"
    }

    resource_type = "EC2"
  }
}
`, rName)
}

func testAccPlanConfig_advancedSettingUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }
  }

  advanced_backup_setting {
    backup_options = {
      WindowsVSS = "disabled"
    }

    resource_type = "EC2"
  }
}
`, rName)
}

func testAccPlanConfig_enableContinuous(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name                = %[1]q
    target_vault_name        = aws_backup_vault.test.name
    schedule                 = "cron(0 12 * * ? *)"
    enable_continuous_backup = true

    lifecycle {
      delete_after = 35
    }
  }
}
`, rName)
}

func testAccPlanConfig_scheduleExpressionTimezone(rName, scheduleExpressionTimezone string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name                    = %[1]q
    target_vault_name            = aws_backup_vault.test.name
    schedule                     = "cron(0 12 * * ? *)"
    schedule_expression_timezone = %[2]q
  }
}
`, rName, scheduleExpressionTimezone)
}
