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

func TestAccWorkMailDefaultDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_default_domain.test"
	domainName := fmt.Sprintf("%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultDomainConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultDomainExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
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

func TestAccWorkMailDefaultDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_default_domain.test"
	domainName := fmt.Sprintf("%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultDomainConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultDomainExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkmail.ResourceDefaultDomain, resourceName),
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

func TestAccWorkMailDefaultDomain_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_default_domain.test"
	domainName1 := fmt.Sprintf("%s-1.example.com", rName)
	domainName2 := fmt.Sprintf("%s-2.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultDomainConfig_basic(rName, domainName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultDomainExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName1),
				),
			},
			{
				Config: testAccDefaultDomainConfig_twoDomains(rName, domainName1, domainName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultDomainExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName2),
				),
			},
		},
	})
}

func testAccCheckDefaultDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_default_domain" {
				continue
			}

			orgID := rs.Primary.Attributes["organization_id"]

			defaultDomain, err := tfworkmail.FindDefaultDomainByOrgID(ctx, conn, orgID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			// Destroy means returning to test domain. If the default domain is still
			// the custom domain we set, it hasn't been properly destroyed.
			domainName := rs.Primary.Attributes[names.AttrDomainName]
			if defaultDomain == domainName {
				return fmt.Errorf("WorkMail Default Domain %s still set for organization %s", domainName, orgID)
			}
		}

		return nil
	}
}

func testAccCheckDefaultDomainExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDefaultDomain, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDefaultDomain, name, errors.New("not set"))
		}

		orgID := rs.Primary.Attributes["organization_id"]

		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		_, err := tfworkmail.FindDefaultDomainByOrgID(ctx, conn, orgID)
		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameDefaultDomain, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccDefaultDomainConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = %[2]q
}

resource "aws_workmail_default_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = aws_workmail_domain.test.domain_name
}
`, rName, domainName)
}

func testAccDefaultDomainConfig_twoDomains(rName, domainName1, domainName2 string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = %[2]q
}

resource "aws_workmail_domain" "test2" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = %[3]q
}

resource "aws_workmail_default_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = aws_workmail_domain.test2.domain_name
}
`, rName, domainName1, domainName2)
}
