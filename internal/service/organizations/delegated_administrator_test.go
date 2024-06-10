// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegatedAdministrator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDelegatedAdministratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceIdentity, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "delegation_enabled_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "service_principal", servicePrincipal),
				),
			},
		},
	})
}

func testAccDelegatedAdministrator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDelegatedAdministratorDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, resourceName, &organization),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceDelegatedAdministrator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegatedAdministratorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_delegated_administrator" {
				continue
			}

			_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["service_principal"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Organizations Delegated Administrator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDelegatedAdministratorExists(ctx context.Context, n string, v *awstypes.DelegatedAdministrator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		output, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["service_principal"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDelegatedAdministratorConfig_basic(servicePrincipal string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}
`, servicePrincipal))
}
