// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID := "111111111111"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(accountID, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEmail), knownvalue.StringExact(acctest.DefaultEmailAddress)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Created")),
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
	resourceName := "aws_guardduty_member.test"
	accountID := "111111111111"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(accountID, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfguardduty.ResourceMember(), resourceName),
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

func testAccMember_invite_disassociate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccMemberFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Invited")),
				},
			},
			// Disassociate member
			{
				Config: testAccMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Removed")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
				},
			},
		},
	})
}

func testAccMember_invite_onUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccMemberFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Created")),
				},
			},
			// Invite member
			{
				Config: testAccMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Invited")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
				},
			},
		},
	})
}

func testAccMember_invitationMessage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccMemberFromEnv(t)
	invitationMessage := "inviting"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invitationMessage(accountID, email, invitationMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("disable_email_notification"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEmail), knownvalue.StringExact(email)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invitation_message"), knownvalue.StringExact(invitationMessage)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invite"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("relationship_status"), knownvalue.StringExact("Invited")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
					"invitation_message",
				},
			},
		},
	})
}

func testAccCheckMemberDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_member" {
				continue
			}

			_, err := tfguardduty.FindMemberByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrAccountID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GuardDuty Member %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMemberExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindMemberByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrAccountID])

		return err
	}
}

func testAccMemberConfig_basic(accountID, email string) string {
	return acctest.ConfigCompose(testAccDetectorConfig_basic, fmt.Sprintf(`
resource "aws_guardduty_member" "test" {
  account_id  = %[1]q
  detector_id = aws_guardduty_detector.test.id
  email       = %[2]q
}
`, accountID, email))
}

func testAccMemberConfig_invite(accountID, email string, invite bool) string {
	return acctest.ConfigCompose(testAccDetectorConfig_basic, fmt.Sprintf(`
resource "aws_guardduty_member" "test" {
  account_id                 = %[1]q
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = %[2]q
  invite                     = %[3]t
}
`, accountID, email, invite))
}

func testAccMemberConfig_invitationMessage(accountID, email, invitationMessage string) string {
	return acctest.ConfigCompose(testAccDetectorConfig_basic, fmt.Sprintf(`
resource "aws_guardduty_member" "test" {
  account_id                 = %[1]q
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = %[2]q
  invitation_message         = %[3]q
  invite                     = true
}
`, accountID, email, invitationMessage))
}
