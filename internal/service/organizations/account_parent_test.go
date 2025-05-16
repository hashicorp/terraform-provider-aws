// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOrganizationsAccountParent_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_organizations_account_parent.test"
	ou1Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ou2Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ou3Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountParentConfig_organizationalUnit1(ou1Name, ou2Name, ou3Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParentID), knownvalue.StringRegexp(regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}$`))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
			},
			{
				Config: testAccAccountParentConfig_organizationalUnit2(ou1Name, ou2Name, ou3Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParentID), knownvalue.StringRegexp(regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}$`))),
				},
			},
			{
				Config: testAccAccountParentConfig_organizationalUnit3(ou1Name, ou2Name, ou3Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParentID), knownvalue.StringRegexp(regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}$`))),
				},
			},
			{
				Config: testAccAccountParentConfig_root(ou1Name, ou2Name, ou3Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParentID), knownvalue.StringRegexp(regexache.MustCompile(`^r-[0-9a-z]{4,32}$`))),
				},
			},
		},
	})
}

func TestAccOrganizationsAccountParent_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_organizations_account_parent.test"
	ou1Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ou2Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ou3Name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountParentConfig_root(ou1Name, ou2Name, ou3Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceAccountParent, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountParentExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAccountParent, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrParentID] == "" {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAccountParent, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		_, err := tforganizations.FindParentAccountID(ctx, conn, rs.Primary.Attributes[names.AttrAccountID])
		if err != nil {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAccountParent, rs.Primary.Attributes[names.AttrAccountID], err)
		}

		return nil
	}
}

func testAccAccountParentConfig_base(firstOU, secondOU, thirdOU string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "first" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_organizations_organizational_unit" "second" {
  name      = %[2]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_organizations_organizational_unit" "third" {
  name      = %[3]q
  parent_id = aws_organizations_organizational_unit.second.id
}

`, firstOU, secondOU, thirdOU))
}

func testAccAccountParentConfig_root(firstOU, secondOU, thirdOU string) string {
	return acctest.ConfigCompose(testAccAccountParentConfig_base(firstOU, secondOU, thirdOU), `
resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = data.aws_organizations_organization.current.roots[0].id
}
`)
}

func testAccAccountParentConfig_organizationalUnit1(firstOU, secondOU, thirdOU string) string {
	return acctest.ConfigCompose(testAccAccountParentConfig_base(firstOU, secondOU, thirdOU), `
resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = aws_organizations_organizational_unit.first.id
}
`)
}

func testAccAccountParentConfig_organizationalUnit2(firstOU, secondOU, thirdOU string) string {
	return acctest.ConfigCompose(testAccAccountParentConfig_base(firstOU, secondOU, thirdOU), `
resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = aws_organizations_organizational_unit.second.id
}
`)
}

func testAccAccountParentConfig_organizationalUnit3(firstOU, secondOU, thirdOU string) string {
	return acctest.ConfigCompose(testAccAccountParentConfig_base(firstOU, secondOU, thirdOU), `
resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = aws_organizations_organizational_unit.third.id
}
`)
}
