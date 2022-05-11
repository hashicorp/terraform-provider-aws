package backup_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
)

func TestAccBackupPlan_basic(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexp.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rName,
						"target_vault_name": rName,
						"schedule":          "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccBackupPlan_withTags(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccBackupPlan_withRules(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))
	rule1Name := fmt.Sprintf("%s_1", rName)
	rule2Name := fmt.Sprintf("%s_2", rName)
	rule3Name := fmt.Sprintf("%s_3", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_twoRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule1Name,
						"target_vault_name": rName,
						"schedule":          "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule2Name,
						"target_vault_name": rName,
						"schedule":          "cron(0 6 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule1Name,
						"target_vault_name": rName,
						"schedule":          "cron(0 6 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule2Name,
						"target_vault_name": rName,
						"schedule":          "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rule3Name,
						"target_vault_name": rName,
						"schedule":          "cron(0 18 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rName,
						"target_vault_name": rName,
						"schedule":          "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccBackupPlan_withLifecycle(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_lifecycleColdStorageAfterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                      rName,
						"lifecycle.#":                    "1",
						"lifecycle.0.cold_storage_after": "30",
						"lifecycle.0.delete_after":       "180",
					}),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_recoveryPointTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						"schedule":                 "cron(0 12 * * ? *)",
						"lifecycle.#":              "0",
						"recovery_point_tags.%":    "3",
						"recovery_point_tags.Name": rName,
						"recovery_point_tags.Key1": "Value1",
						"recovery_point_tags.Key2": "Value2a",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						"schedule":                 "cron(0 12 * * ? *)",
						"lifecycle.#":              "0",
						"recovery_point_tags.%":    "3",
						"recovery_point_tags.Name": rName,
						"recovery_point_tags.Key2": "Value2b",
						"recovery_point_tags.Key3": "Value3",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":         rName,
						"target_vault_name": rName,
						"schedule":          "cron(0 12 * * ? *)",
						"lifecycle.#":       "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccBackupPlan_RuleCopyAction_sameRegion(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanRuleCopyActionConfig(rName, 30, 180),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
				Config: testAccPlanRuleCopyActionConfig(rName, 60, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanRuleCopyActionNoLifecycleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
				Config: testAccPlanRuleCopyActionConfig(rName, 60, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
				Config: testAccPlanRuleCopyActionNoLifecycleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanRuleCopyActionMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	var providers []*schema.Provider
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanRuleCopyActionCrossRegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
				Config:            testAccPlanRuleCopyActionCrossRegionConfig(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupPlan_advancedBackupSetting(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanAdvancedBackupSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "advanced_backup_setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_backup_setting.*", map[string]string{
						"backup_options.%":          "1",
						"backup_options.WindowsVSS": "enabled",
						"resource_type":             "EC2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanAdvancedBackupSettingUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "advanced_backup_setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_backup_setting.*", map[string]string{
						"backup_options.%":          "1",
						"backup_options.WindowsVSS": "disabled",
						"resource_type":             "EC2",
					}),
				),
			},
		},
	})
}

func TestAccBackupPlan_enableContinuousBackup(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanEnableContinuousBackupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexp.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_name":                rName,
						"target_vault_name":        rName,
						"schedule":                 "cron(0 12 * * ? *)",
						"enable_continuous_backup": "true",
						"lifecycle.#":              "1",
						"lifecycle.0.delete_after": "35",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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
	var plan backup.GetBackupPlanOutput
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(resourceName, &plan),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourcePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPlanDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_plan" {
			continue
		}

		input := &backup.GetBackupPlanInput{
			BackupPlanId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupPlan(input)

		if err == nil {
			if *resp.BackupPlanId == rs.Primary.ID {
				return fmt.Errorf("Plane '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckPlanExists(name string, plan *backup.GetBackupPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := conn.GetBackupPlan(&backup.GetBackupPlanInput{
			BackupPlanId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*plan = *output

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

func testAccPlanConfig_tags(rName string) string {
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

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccPlanConfig_tagsUpdated(rName string) string {
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

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
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

func testAccPlanRuleCopyActionConfig(rName string, coldStorageAfter, deleteAfter int) string {
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

func testAccPlanRuleCopyActionMultipleConfig(rName string) string {
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

func testAccPlanRuleCopyActionCrossRegionConfig(rName string) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
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
`, rName)
}

func testAccPlanRuleCopyActionNoLifecycleConfig(rName string) string {
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

func testAccPlanAdvancedBackupSettingConfig(rName string) string {
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

func testAccPlanAdvancedBackupSettingUpdatedConfig(rName string) string {
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

func testAccPlanEnableContinuousBackupConfig(rName string) string {
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
