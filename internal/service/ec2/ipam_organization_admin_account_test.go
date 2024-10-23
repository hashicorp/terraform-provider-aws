// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	organizationstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMOrganizationAdminAccount_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccIPAMOrganizationAdminAccount_basic,
		acctest.CtDisappears: testAccIPAMOrganizationAdminAccount_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

// Prerequisites:
// * Organizations management account
// * Organization member account
// Authenticate with management account as target account and member account as alternate.
func testAccIPAMOrganizationAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var organization organizationstypes.DelegatedAdministrator
	resourceName := "aws_vpc_ipam_organization_admin_account.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckIPAMOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers.
				Config: testAccIPAMOrganizationAdminAccountConfig_init,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationMemberAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccIPAMOrganizationAdminAccountConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMOrganizationAdminAccountExists(ctx, resourceName, &organization),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "organizations", regexache.MustCompile("account/.+")),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceIdentity, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "service_principal", tfec2.IPAMServicePrincipal),
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

func testAccIPAMOrganizationAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var organization organizationstypes.DelegatedAdministrator
	resourceName := "aws_vpc_ipam_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckIPAMOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers.
				Config: testAccIPAMOrganizationAdminAccountConfig_init,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationMemberAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccIPAMOrganizationAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMOrganizationAdminAccountExists(ctx, resourceName, &organization),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMOrganizationAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIPAMOrganizationAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_organization_admin_account" {
				continue
			}

			_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.Attributes["delegated_admin_account_id"], rs.Primary.Attributes["service_principal"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Organization Delegated Admin Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPAMOrganizationAdminAccountExists(ctx context.Context, n string, v *organizationstypes.DelegatedAdministrator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		output, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.Attributes["delegated_admin_account_id"], rs.Primary.Attributes["service_principal"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// Initialize all the providers used by organization administrator acceptance tests.
var testAccIPAMOrganizationAdminAccountConfig_init = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}
`)

var testAccIPAMOrganizationAdminAccountConfig_basic = acctest.ConfigCompose(testAccIPAMOrganizationAdminAccountConfig_init, `
resource "aws_vpc_ipam_organization_admin_account" "test" {
  delegated_admin_account_id = data.aws_caller_identity.delegated.account_id
}
`)
