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
	"github.com/hashicorp/terraform-provider-aws/names"

	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

func TestAccOrganizationsAwsServiceAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var awsserviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsServiceAccessExists(ctx, t, resourceName, &awsserviceaccess),
					resource.TestCheckResourceAttr(resourceName, "service_principal", servicePrincipal),
					resource.TestCheckResourceAttrSet(resourceName, "date_enabled"),
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

func TestAccOrganizationsAwsServiceAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var awsserviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsServiceAccessExists(ctx, t, resourceName, &awsserviceaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tforganizations.ResourceAwsServiceAccess, resourceName),
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

func testAccCheckAwsServiceAccessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_aws_service_access" {
				continue
			}

			servicePrincipal := rs.Primary.Attributes["service_principal"]
			_, err := tforganizations.FindAwsServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)
			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameAwsServiceAccess, servicePrincipal, err)
			}

			return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameAwsServiceAccess, servicePrincipal, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAwsServiceAccessExists(ctx context.Context, t *testing.T, name string, awsserviceaccess *awstypes.EnabledServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, name, errors.New("not found"))
		}

		servicePrincipal := rs.Primary.Attributes["service_principal"]
		if servicePrincipal == "" {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)
		resp, err := tforganizations.FindAwsServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)
		if err != nil {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, servicePrincipal, err)
		}

		*awsserviceaccess = *resp

		return nil
	}
}

func testAccAwsServiceAccessConfig_basic(servicePrincipal string) string {
	return fmt.Sprintf(`
resource "aws_organizations_aws_service_access" "test" {
  service_principal             = %[1]q
}
`, servicePrincipal)
}
