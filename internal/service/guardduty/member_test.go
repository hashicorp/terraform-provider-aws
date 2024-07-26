// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID := "111111111111"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(accountID, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Created"),
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

func testAccMember_invite_disassociate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
			},
			// Disassociate member
			{
				Config: testAccMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Removed"),
				),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Created"),
				),
			},
			// Invite member
			{
				Config: testAccMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invitationMessage(accountID, email, invitationMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "disable_email_notification", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "invitation_message", invitationMessage),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
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

func testAccCheckMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_member" {
				continue
			}

			accountID, detectorID, err := tfguardduty.DecodeMemberID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &guardduty.GetMembersInput{
				AccountIds: []*string{aws.String(accountID)},
				DetectorId: aws.String(detectorID),
			}

			gmo, err := conn.GetMembersWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
					return nil
				}
				return err
			}

			if len(gmo.Members) < 1 {
				continue
			}

			return fmt.Errorf("Expected GuardDuty Detector to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMemberExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		accountID, detectorID, err := tfguardduty.DecodeMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetMembersInput{
			AccountIds: []*string{aws.String(accountID)},
			DetectorId: aws.String(detectorID),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)
		gmo, err := conn.GetMembersWithContext(ctx, input)
		if err != nil {
			return err
		}

		if len(gmo.Members) < 1 {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccMemberConfig_basic(accountID, email string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id  = "%[2]s"
  detector_id = aws_guardduty_detector.test.id
  email       = "%[3]s"
}
`, testAccDetectorConfig_basic, accountID, email)
}

func testAccMemberConfig_invite(accountID, email string, invite bool) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id                 = "%[2]s"
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = "%[3]s"
  invite                     = %[4]t
}
`, testAccDetectorConfig_basic, accountID, email, invite)
}

func testAccMemberConfig_invitationMessage(accountID, email, invitationMessage string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id                 = "%[2]s"
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = "%[3]s"
  invitation_message         = "%[4]s"
  invite                     = true
}
`, testAccDetectorConfig_basic, accountID, email, invitationMessage)
}
