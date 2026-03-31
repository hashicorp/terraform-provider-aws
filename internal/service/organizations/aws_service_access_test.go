// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationsAWSServiceAccess_basic(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	var serviceaccess awstypes.EnabledServicePrincipal
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AWSServiceAccess/basic/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSServiceAccessExists(ctx, t, resourceName, &serviceaccess),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("date_enabled"), knownvalue.NotNull()),
				},
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

func testAccOrganizationsAWSServiceAccess_disappears(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	var serviceaccess awstypes.EnabledServicePrincipal
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AWSServiceAccess/basic/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSServiceAccessExists(ctx, t, resourceName, &serviceaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tforganizations.ResourceAWSServiceAccess, resourceName),
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

func testAccCheckAWSServiceAccessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc { // nosemgrep:ci.aws-in-func-name
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_aws_service_access" {
				continue
			}

			_, err := tforganizations.FindAWSServiceAccessByServicePrincipal(ctx, conn, rs.Primary.Attributes["service_principal"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Organizations AWS Service Access %s still exists", rs.Primary.Attributes["service_principal"])
		}

		return nil
	}
}

func testAccCheckAWSServiceAccessExists(ctx context.Context, t *testing.T, n string, v *awstypes.EnabledServicePrincipal) resource.TestCheckFunc { // nosemgrep:ci.aws-in-func-name
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		output, err := tforganizations.FindAWSServiceAccessByServicePrincipal(ctx, conn, rs.Primary.Attributes["service_principal"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
