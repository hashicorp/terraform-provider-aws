// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic("111111111111"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.StringExact("111111111111")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEmail), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("member_status"), knownvalue.StringExact("Created")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMember_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic("111111111111"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecurityhub.ResourceMember(), resourceName),
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

func testAccMember_inviteTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite("111111111111", acctest.DefaultEmailAddress, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.StringExact("111111111111")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEmail), knownvalue.StringExact(acctest.DefaultEmailAddress)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("member_status"), knownvalue.StringExact("Invited")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"invite"},
			},
		},
	})
}

func testAccMember_inviteFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite("111111111111", acctest.DefaultEmailAddress, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), knownvalue.StringExact("111111111111")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEmail), knownvalue.StringExact(acctest.DefaultEmailAddress)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("member_status"), knownvalue.StringExact("Created")),
				},
			},
			{
				// Re-apply the same configuration to verify no drift/replacement.
				Config: testAccMemberConfig_invite("111111111111", acctest.DefaultEmailAddress, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMember_inviteOrganizationMember(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var member types.Member
	resourceName := "aws_securityhub_member.test"
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName(t))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers.
				Config: testAccMemberOrganizationMemberConfig_init,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationMemberAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccMemberConfig_inviteOrganizationMember(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName, &member),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMemberExists(ctx context.Context, t *testing.T, n string, v *types.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMemberDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_member" {
				continue
			}

			_, err := tfsecurityhub.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Member %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMemberConfig_basic(accountID string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = %[1]q
}
`, accountID)
}

func testAccMemberConfig_invite(accountID, email string, invite bool) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = %[1]q
  email      = %[2]q
  invite     = %[3]t
}
`, accountID, email, invite)
}

// Initialize all the providers used by organization member acceptance tests.
var testAccMemberOrganizationMemberConfig_init = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}
`)

func testAccMemberConfig_inviteOrganizationMember(email string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = data.aws_caller_identity.member.account_id
  email      = %[1]q
  invite     = false
}
`, email))
}
