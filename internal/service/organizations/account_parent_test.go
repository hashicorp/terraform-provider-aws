// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

func TestAccOrganizationsAccountParent_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_organizations_account_parent.test"
	ouName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAccountParentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountParentConfig_root(),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParentID), knownvalue.StringRegexp(regexache.MustCompile(`^r-[0-9a-z]{4,32}$`))),
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
				Config: testAccAccountParentConfig_organizationalUnit(ouName),
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
		},
	})
}

func TestAccOrganizationsAccountParent_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_organizations_account_parent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAccountParentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountParentConfig_root(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountParentExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceAccountParent, resourceName),
				),
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

func testAccCheckAccountParentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		root, err := tforganizations.FindDefaultRoot(ctx, conn)
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_account_parent" {
				continue
			}

			parentID, err := tforganizations.FindParentAccountID(ctx, conn, rs.Primary.Attributes[names.AttrAccountID])
			if err != nil {
				return err
			}

			if *parentID == *root.Id {
				continue
			}

			return fmt.Errorf("Found account %s outside of org root", rs.Primary.Attributes[names.AttrAccountID])
		}
		return nil
	}
}

func testAccAccountParentConfig_root() string {
	return `
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = data.aws_organizations_organization.current.roots[0].id
}
`
}

func testAccAccountParentConfig_organizationalUnit(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_organizations_account_parent" "test" {
  account_id = data.aws_caller_identity.alternate.account_id
  parent_id  = aws_organizations_organizational_unit.test.id
}
`, rName)
}
