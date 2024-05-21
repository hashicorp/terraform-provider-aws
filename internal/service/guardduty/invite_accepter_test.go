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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInviteAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	masterDetectorResourceName := "aws_guardduty_detector.master"
	memberDetectorResourceName := "aws_guardduty_detector.member"
	resourceName := "aws_guardduty_invite_accepter.test"
	_, email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInviteAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInviteAccepterConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInviteAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", memberDetectorResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_id", masterDetectorResourceName, names.AttrAccountID),
				),
			},
			{
				Config:            testAccInviteAccepterConfig_basic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInviteAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_invite_accepter" {
				continue
			}

			input := &guardduty.GetMasterAccountInput{
				DetectorId: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetMasterAccountWithContext(ctx, input)

			if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}

			if err != nil {
				return err
			}

			if output == nil || output.Master == nil || aws.StringValue(output.Master.AccountId) != rs.Primary.Attributes["master_account_id"] {
				continue
			}

			return fmt.Errorf("GuardDuty Detector (%s) still has GuardDuty Master Account ID (%s)", rs.Primary.ID, aws.StringValue(output.Master.AccountId))
		}

		return nil
	}
}

func testAccCheckInviteAccepterExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) has empty ID", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		input := &guardduty.GetMasterAccountInput{
			DetectorId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetMasterAccountWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || output.Master == nil || aws.StringValue(output.Master.AccountId) == "" {
			return fmt.Errorf("no master account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccInviteAccepterConfig_basic(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
resource "aws_guardduty_detector" "master" {
  provider = "awsalternate"
}

resource "aws_guardduty_detector" "member" {}

resource "aws_guardduty_member" "member" {
  provider = "awsalternate"

  account_id                 = aws_guardduty_detector.member.account_id
  detector_id                = aws_guardduty_detector.master.id
  disable_email_notification = true
  email                      = %q
  invite                     = true
}

resource "aws_guardduty_invite_accepter" "test" {
  depends_on = [aws_guardduty_member.member]

  detector_id       = aws_guardduty_detector.member.id
  master_account_id = aws_guardduty_detector.master.account_id
}
`, email)
}
