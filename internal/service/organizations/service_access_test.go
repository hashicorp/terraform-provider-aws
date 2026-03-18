// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOrganizationsServiceAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var serviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceAccessExists(ctx, t, resourceName, &serviceaccess),
					resource.TestCheckResourceAttr(resourceName, "service_principal", servicePrincipal),
					resource.TestCheckResourceAttrSet(resourceName, "date_enabled"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "service_principal",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "service_principal"),
			},
		},
	})
}

func TestAccOrganizationsServiceAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var serviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceAccessExists(ctx, t, resourceName, &serviceaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tforganizations.ResourceServiceAccess, resourceName),
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

func testAccCheckServiceAccessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_service_access" {
				continue
			}

			servicePrincipal := rs.Primary.Attributes["service_principal"]
			_, err := tforganizations.FindServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)
			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameServiceAccess, servicePrincipal, err)
			}

			return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameServiceAccess, servicePrincipal, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckServiceAccessExists(ctx context.Context, t *testing.T, name string, serviceaccess *awstypes.EnabledServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameServiceAccess, name, errors.New("not found"))
		}

		servicePrincipal := rs.Primary.Attributes["service_principal"]
		if servicePrincipal == "" {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameServiceAccess, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)
		resp, err := tforganizations.FindServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)
		if err != nil {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameServiceAccess, servicePrincipal, err)
		}

		*serviceaccess = *resp

		return nil
	}
}

func testAccServiceAccessConfig_basic(servicePrincipal string) string {
	return fmt.Sprintf(`
resource "aws_organizations_service_access" "test" {
  service_principal = %[1]q
}
`, servicePrincipal)
}
