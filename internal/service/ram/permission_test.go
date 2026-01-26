// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var permission awstypes.ResourceSharePermissionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ram", fmt.Sprintf("permission/%s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("default_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"policy_template"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccRAMPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var permission awstypes.ResourceSharePermissionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfram.ResourcePermission, resourceName),
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

func TestAccRAMPermission_version(t *testing.T) {
	ctx := acctest.Context(t)
	var permission awstypes.ResourceSharePermissionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"
	policyTemplateActions := []string{
		"backup:ListProtectedResourcesByBackupVault",
		"backup:ListRecoveryPointsByBackupVault",
		"backup:DescribeRecoveryPoint",
		"backup:DescribeBackupVault",
		"backup:StartRestoreJob",
	}
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[:3], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"policy_template"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[1:2], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[3:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[2:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions, `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
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

func TestAccRAMPermission_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var permission awstypes.ResourceSharePermissionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"policy_template"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPermissionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccPermissionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName, &permission),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckPermissionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_permission" {
				continue
			}

			_, err := tfram.FindPermissionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Permission %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckPermissionExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceSharePermissionDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		output, err := tfram.FindPermissionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

	input := &ram.ListPermissionsInput{}

	_, err := conn.ListPermissions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPermissionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_permission" "test" {
  name            = %[1]q
  policy_template = <<EOF
{
    "Effect": "Allow",
    "Action": [
	"backup:ListProtectedResourcesByBackupVault",
	"backup:ListRecoveryPointsByBackupVault",
	"backup:DescribeRecoveryPoint",
	"backup:DescribeBackupVault"
    ]
}
EOF
  resource_type   = "backup:BackupVault"
}
`, rName)
}

func testAccPermissionConfig_version(rName, rPolicyTemplate string) string {
	return fmt.Sprintf(`
resource "aws_ram_permission" "test" {
  name            = %[1]q
  policy_template = <<EOF
{
	"Effect": "Allow",
	"Action": [
	     %[2]s
	]
}
EOF
  resource_type   = "backup:BackupVault"

  tags = {
    Name = %[1]q
  }
}
`, rName, rPolicyTemplate)
}

func testAccPermissionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_permission" "test" {
  name            = %[1]q
  policy_template = <<EOF
{
    "Effect": "Allow",
    "Action": [
	"backup:ListProtectedResourcesByBackupVault",
	"backup:ListRecoveryPointsByBackupVault",
	"backup:DescribeRecoveryPoint",
	"backup:DescribeBackupVault"
    ]
}
EOF
  resource_type   = "backup:BackupVault"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPermissionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ram_permission" "test" {
  name            = %[1]q
  policy_template = <<EOF
{
    "Effect": "Allow",
    "Action": [
	"backup:ListProtectedResourcesByBackupVault",
	"backup:ListRecoveryPointsByBackupVault",
	"backup:DescribeRecoveryPoint",
	"backup:DescribeBackupVault"
    ]
}
EOF
  resource_type   = "backup:BackupVault"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
