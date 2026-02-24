// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegatedAdministrator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "securitylake.amazonaws.com"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDelegatedAdministratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, t, resourceName, &organization),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile(`account/o-[0-9a-z]{10}/\d{12}$`)),
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
	servicePrincipal := "securitylake.amazonaws.com"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDelegatedAdministratorDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, t, resourceName, &organization),
					acctest.CheckSDKResourceDisappears(ctx, t, tforganizations.ResourceDelegatedAdministrator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationsDelegatedAdministrator_identitySerial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:    testAccOrganizationsDelegatedAdministrator_Identity_basic,
		"ExistingResource": testAccOrganizationsDelegatedAdministrator_Identity_ExistingResource_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOrganizationsDelegatedAdministrator_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "securitylake.amazonaws.com"

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDelegatedAdministratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectAttributeFormat(resourceName, tfjsonpath.New(names.AttrID), "{account_id}/{service_principal}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:    tfknownvalue.AccountID(),
						"service_principal":    knownvalue.NotNull(),
						"delegated_account_id": knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("service_principal")),
					statecheck.ExpectIdentityValueMatchesStateAtPath(resourceName, tfjsonpath.New("delegated_account_id"), tfjsonpath.New(names.AttrAccountID)),
				},
			},

			// Step 2: Import command
			{
				Config:            testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_principal"), knownvalue.NotNull()),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_principal"), knownvalue.NotNull()),
					},
				},
			},
		},
	})
}

func testAccOrganizationsDelegatedAdministrator_Identity_ExistingResource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "securitylake.amazonaws.com"
	providers := make(map[string]*schema.Provider)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.OrganizationsServiceID),
		CheckDestroy: testAccCheckDelegatedAdministratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.4.0",
					},
				},
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers, acctest.ProviderNameAlternate),
				Config:                   testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDelegatedAdministratorExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
				Config:                   testAccDelegatedAdministratorConfig_basic(servicePrincipal),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:    tfknownvalue.AccountID(),
						"service_principal":    knownvalue.NotNull(),
						"delegated_account_id": knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("service_principal")),
					statecheck.ExpectIdentityValueMatchesStateAtPath(resourceName, tfjsonpath.New("delegated_account_id"), tfjsonpath.New(names.AttrAccountID)),
				},
			},
		},
	})
}

func testAccCheckDelegatedAdministratorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_delegated_administrator" {
				continue
			}

			_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["service_principal"])

			if retry.NotFound(err) {
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

func testAccCheckDelegatedAdministratorExists(ctx context.Context, t *testing.T, n string, v *awstypes.DelegatedAdministrator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

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
