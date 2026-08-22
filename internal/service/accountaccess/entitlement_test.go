// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfaccountaccess "github.com/hashicorp/terraform-provider-aws/internal/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessEntitlement_user(t *testing.T) {
	ctx := acctest.Context(t)

	var v accountaccess.GetEntitlementOutput
	resourceName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_user(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "entitlement_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// account_name is server-resolved with eventual consistency
				// (empty immediately after create, populated by import time).
				ImportStateVerifyIgnore: []string{"account_name"},
			},
		},
	})
}

func testAccAccountAccessEntitlement_group(t *testing.T) {
	ctx := acctest.Context(t)

	var v accountaccess.GetEntitlementOutput
	resourceName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_group(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// account_name is server-resolved with eventual consistency
				// (empty immediately after create, populated by import time).
				ImportStateVerifyIgnore: []string{"account_name"},
			},
		},
	})
}

func testAccAccountAccessEntitlement_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v accountaccess.GetEntitlementOutput
	resourceName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementConfig_user(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntitlementExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfaccountaccess.ResourceEntitlement, resourceName),
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

func testAccCheckEntitlementDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accountaccess_entitlement" {
				continue
			}

			applicationArn := rs.Primary.Attributes["application_arn"]
			entitlementID := rs.Primary.Attributes["entitlement_id"]

			_, err := tfaccountaccess.FindEntitlementByTwoPartKey(ctx, conn, applicationArn, entitlementID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Account Access Entitlement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEntitlementExists(ctx context.Context, t *testing.T, n string, v *accountaccess.GetEntitlementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Account Access Entitlement ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		applicationArn := rs.Primary.Attributes["application_arn"]
		entitlementID := rs.Primary.Attributes["entitlement_id"]

		output, err := tfaccountaccess.FindEntitlementByTwoPartKey(ctx, conn, applicationArn, entitlementID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEntitlementConfig_user(rName string) string {
	return acctest.ConfigCompose(testAccPrerequisitesConfig(rName), `
resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = local.instance_arn
}

resource "aws_accountaccess_entitlement" "test" {
  application_arn = aws_accountaccess_application.test.arn
  principal_id    = aws_identitystore_user.test.user_id
  principal_type  = "USER"
  role_arn        = aws_iam_role.test.arn
}
`)
}

func testAccEntitlementConfig_group(rName string) string {
	return acctest.ConfigCompose(testAccPrerequisitesConfig(rName), `
resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = local.instance_arn
}

resource "aws_accountaccess_entitlement" "test" {
  application_arn = aws_accountaccess_application.test.arn
  principal_id    = aws_identitystore_group.test.group_id
  principal_type  = "GROUP"
  role_arn        = aws_iam_role.test.arn
}
`)
}
