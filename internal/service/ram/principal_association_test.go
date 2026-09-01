// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMPrincipalAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_baseIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Prevent "waiting for RAM Resource Share (...) Principal (...) Association create: unexpected state 'FAILED', wanted target 'ASSOCIATED'".
					acctest.CheckSleep(t, 1*time.Minute),
				),
			},
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, t, resourceName, &association),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
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

func TestAccRAMPrincipalAssociation_accountID(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, t, resourceName, &association),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
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

func TestAccRAMPrincipalAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_baseIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Prevent "waiting for RAM Resource Share (...) Principal (...) Association create: unexpected state 'FAILED', wanted target 'ASSOCIATED'".
					acctest.CheckSleep(t, 1*time.Minute),
				),
			},
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, t, resourceName, &association),
					acctest.CheckSDKResourceDisappears(ctx, t, tfram.ResourcePrincipalAssociation(), resourceName),
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

func TestAccRAMPrincipalAssociation_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_baseIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Prevent "waiting for RAM Resource Share (...) Principal (...) Association create: unexpected state 'FAILED', wanted target 'ASSOCIATED'".
					acctest.CheckSleep(t, 1*time.Minute),
				),
			},
			{
				Config:      testAccPrincipalAssociationConfig_duplicate(rName),
				ExpectError: regexache.MustCompile(`RAM Principal Association .* already exists`),
			},
		},
	})
}

func testAccCheckPrincipalAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		output, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes[names.AttrPrincipal])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPrincipalAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_principal_association" {
				continue
			}

			_, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes[names.AttrPrincipal])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Principal Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPrincipalAssociationConfig_baseIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = false
  name                      = %[1]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
    }]
  })
}
`, rName)
}

func testAccPrincipalAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPrincipalAssociationConfig_baseIAMRole(rName), `
resource "aws_ram_principal_association" "test" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`)
}

func testAccPrincipalAssociationConfig_accountID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName))
}

func testAccPrincipalAssociationConfig_duplicate(rName string) string {
	return acctest.ConfigCompose(testAccPrincipalAssociationConfig_baseIAMRole(rName), `
resource "aws_ram_principal_association" "test1" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test2" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_principal_association.test1.resource_share_arn
}
`)
}
