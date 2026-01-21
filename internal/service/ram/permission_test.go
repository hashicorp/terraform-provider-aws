// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var permission awstypes.ResourceSharePermissionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "default_version"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
		},
	})
}

func TestAccRAMPermission_version(t *testing.T) {
	ctx := acctest.Context(t)

	var permission awstypes.ResourceSharePermissionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"
	policyTemplateActions := []string{
		"backup:ListProtectedResourcesByBackupVault",
		"backup:ListRecoveryPointsByBackupVault",
		"backup:DescribeRecoveryPoint",
		"backup:DescribeBackupVault",
		"backup:StartRestoreJob",
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[:3], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[1:2], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[3:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions[2:4], `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_template"},
			},
			{
				Config: testAccPermissionConfig_version(rName, `"`+strings.Join(policyTemplateActions, `", "`)+`"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ram", "permission/{name}"),
				),
			},
		},
	})
}

func TestAccRAMPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var permission awstypes.ResourceSharePermissionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RAM)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &permission),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfram.ResourcePermission, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckPermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_permission" {
				continue
			}

			output, err := tfram.FindPermissionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if output != nil && output.Status == awstypes.PermissionStatusDeleted {
				return nil
			}

			return fmt.Errorf("RAM Permission %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckPermissionExists(ctx context.Context, name string, permission *awstypes.ResourceSharePermissionDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		resp, err := tfram.FindPermissionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*permission = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

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

  tags = {
    Name = %[1]q
  }
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
