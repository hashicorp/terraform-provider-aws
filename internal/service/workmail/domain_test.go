// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkMailDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_domain.test"
	domainName := fmt.Sprintf("%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dkim_verification_status"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_status"),
					resource.TestCheckResourceAttrSet(resourceName, "is_default"),
					resource.TestCheckResourceAttrSet(resourceName, "is_test_domain"),
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

func TestAccWorkMailDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_domain.test"
	domainName := fmt.Sprintf("%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkmail.ResourceDomain, resourceName),
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

func testAccCheckDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_domain" {
				continue
			}

			orgID := rs.Primary.Attributes["organization_id"]
			domainName := rs.Primary.Attributes[names.AttrDomainName]

			_, err := tfworkmail.FindDomainByOrgAndName(ctx, conn, orgID, domainName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkMail Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDomain, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDomain, name, errors.New("not set"))
		}

		orgID := rs.Primary.Attributes["organization_id"]
		domainName := rs.Primary.Attributes[names.AttrDomainName]

		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		_, err := tfworkmail.FindDomainByOrgAndName(ctx, conn, orgID, domainName)
		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDomain, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccDomainConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = %[2]q
}
`, rName, domainName)
}
