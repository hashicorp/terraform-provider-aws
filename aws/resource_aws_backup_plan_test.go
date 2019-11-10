package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsBackupPlan_basic(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexp.MustCompile(`backup-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withTags(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "50"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withRules(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))
	rule1Name := fmt.Sprintf("%s_1", rName)
	rule2Name := fmt.Sprintf("%s_2", rName)
	rule3Name := fmt.Sprintf("%s_3", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_twoRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "rule_name", rule1Name),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "recovery_point_tags.%", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "rule_name", rule2Name),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "schedule", "cron(0 6 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "recovery_point_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_threeRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "3"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "rule_name", rule1Name),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "schedule", "cron(0 6 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule1Name, "recovery_point_tags.%", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "rule_name", rule2Name),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule2Name, "recovery_point_tags.%", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule3Name, "rule_name", rule3Name),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule3Name, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule3Name, "schedule", "cron(0 18 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule3Name, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rule3Name, "recovery_point_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withLifecycle(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_lifecycleColdStorageAfterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.cold_storage_after", "7"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.delete_after", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_lifecycleDeleteAfterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.cold_storage_after", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.delete_after", "120"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_lifecycleColdStorageAfterAndDeleteAfter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.cold_storage_after", "30"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.0.delete_after", "180"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_withRecoveryPointTags(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_recoveryPointTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "3"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Key1", "Value1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_recoveryPointTagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "3"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Key2", "Value2b"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "rule_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "target_vault_name", rName),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "schedule", "cron(0 12 * * ? *)"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "lifecycle.#", "0"),
					testAccCheckAwsBackupPlanRuleAttr(resourceName, &ruleNameMap, rName, "recovery_point_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsBackupPlan_disappears(t *testing.T) {
	var plan backup.GetBackupPlanOutput
	ruleNameMap := map[string]string{}
	resourceName := "aws_backup_plan.test"
	rName := fmt.Sprintf("tf-testacc-backup-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBackup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupPlanExists(resourceName, &plan, &ruleNameMap),
					testAccCheckAwsBackupPlanDisappears(&plan),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsBackupPlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
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

func testAccCheckAwsBackupPlanDisappears(backupPlan *backup.GetBackupPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).backupconn

		input := &backup.DeleteBackupPlanInput{
			BackupPlanId: backupPlan.BackupPlanId,
		}

		_, err := conn.DeleteBackupPlan(input)

		return err
	}
}

func testAccCheckAwsBackupPlanExists(name string, plan *backup.GetBackupPlanOutput, ruleNameMap *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).backupconn

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

		// Build map of rule name to hash value.
		re := regexp.MustCompile(`^rule\.(\d+)\.rule_name$`)
		for k, v := range rs.Primary.Attributes {
			matches := re.FindStringSubmatch(k)
			if matches != nil {
				(*ruleNameMap)[v] = matches[1]
			}
		}

		return nil
	}
}

func testAccCheckAwsBackupPlanRuleAttr(name string, ruleNameMap *map[string]string, ruleName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(name, fmt.Sprintf("rule.%s.%s", (*ruleNameMap)[ruleName], key), value)(s)
	}
}

func testAccAwsBackupPlanConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
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

func testAccAwsBackupPlanConfig_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"
  }

  // Tag limits:
  //  * Maximum number of tags per resource – 50
  //  * Maximum key length – 128 Unicode characters in UTF-8
  //  * Maximum value length – 256 Unicode characters in UTF-8
  tags = {
    Key___________________________________________________________________________________________________________________________00 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________00"
    Key___________________________________________________________________________________________________________________________01 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________01"
    Key___________________________________________________________________________________________________________________________02 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________02"
    Key___________________________________________________________________________________________________________________________03 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________03"
    Key___________________________________________________________________________________________________________________________04 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________04"
    Key___________________________________________________________________________________________________________________________05 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________05"
    Key___________________________________________________________________________________________________________________________06 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________06"
    Key___________________________________________________________________________________________________________________________07 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________07"
    Key___________________________________________________________________________________________________________________________08 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________08"
    Key___________________________________________________________________________________________________________________________09 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________09"
    Key___________________________________________________________________________________________________________________________10 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________10"
    Key___________________________________________________________________________________________________________________________11 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________11"
    Key___________________________________________________________________________________________________________________________12 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________12"
    Key___________________________________________________________________________________________________________________________13 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________13"
    Key___________________________________________________________________________________________________________________________14 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________14"
    Key___________________________________________________________________________________________________________________________15 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________15"
    Key___________________________________________________________________________________________________________________________16 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________16"
    Key___________________________________________________________________________________________________________________________17 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________17"
    Key___________________________________________________________________________________________________________________________18 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________18"
    Key___________________________________________________________________________________________________________________________19 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________19"
    Key___________________________________________________________________________________________________________________________20 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________20"
    Key___________________________________________________________________________________________________________________________21 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________21"
    Key___________________________________________________________________________________________________________________________22 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________22"
    Key___________________________________________________________________________________________________________________________23 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________23"
    Key___________________________________________________________________________________________________________________________24 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________24"
    Key___________________________________________________________________________________________________________________________25 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________25"
    Key___________________________________________________________________________________________________________________________26 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________26"
    Key___________________________________________________________________________________________________________________________27 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________27"
    Key___________________________________________________________________________________________________________________________28 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________28"
    Key___________________________________________________________________________________________________________________________29 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________29"
    Key___________________________________________________________________________________________________________________________30 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________30"
    Key___________________________________________________________________________________________________________________________31 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________31"
    Key___________________________________________________________________________________________________________________________32 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________32"
    Key___________________________________________________________________________________________________________________________33 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________33"
    Key___________________________________________________________________________________________________________________________34 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________34"
    Key___________________________________________________________________________________________________________________________35 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________35"
    Key___________________________________________________________________________________________________________________________36 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________36"
    Key___________________________________________________________________________________________________________________________37 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________37"
    Key___________________________________________________________________________________________________________________________38 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________38"
    Key___________________________________________________________________________________________________________________________39 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________39"
    Key___________________________________________________________________________________________________________________________40 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________40"
    Key___________________________________________________________________________________________________________________________41 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________41"
    Key___________________________________________________________________________________________________________________________42 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________42"
    Key___________________________________________________________________________________________________________________________43 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________43"
    Key___________________________________________________________________________________________________________________________44 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________44"
    Key___________________________________________________________________________________________________________________________45 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________45"
    Key___________________________________________________________________________________________________________________________46 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________46"
    Key___________________________________________________________________________________________________________________________47 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________47"
    Key___________________________________________________________________________________________________________________________48 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________48"
    Key___________________________________________________________________________________________________________________________49 = "Value_________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________49"
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_twoRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = "%[1]s_1"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_2"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 6 * * ? *)"
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_threeRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = "%[1]s_1"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 6 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_2"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"
  }
  rule {
    rule_name         = "%[1]s_3"
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 18 * * ? *)"
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_lifecycleColdStorageAfterOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 7
    }
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_lifecycleDeleteAfterOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      delete_after = 120
    }
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_lifecycleColdStorageAfterAndDeleteAfter(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      cold_storage_after = 30
      delete_after       = 180
    }
  }
}
`, rName)
}

func testAccAwsBackupPlanConfig_recoveryPointTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
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

func testAccAwsBackupPlanConfig_recoveryPointTagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = "${aws_backup_vault.test.name}"
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
