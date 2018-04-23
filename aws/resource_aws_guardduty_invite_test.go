package aws

import (
	"testing"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

func testACCAwsGuardDutyInvite_basic(t *testing.T) {
	resourceName := "aws_guardduty_invite.test"
	accountID := "111111111111"
	email := "required@example.com"
	message := "foobar"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t)},
		Providers:    testAccProviders,
		CheckDestroy: testACCAwsGuardDutyInviteDestroy,
		Steps: []resource.TestStep {
			{
				Config: testAccGuardDutyInviteConfig_basic(accountID, email, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyInviteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "message", message),
				),
			},
		},
	})
	return
}

func testACCAwsGuardDutyInviteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_invite" {
			continue
		}


		//FIX ME
		accountID, detectorID, message, err := decodeGuardDutyInviteID(rs.Primary.ID)
		if err != nil {
			return nil
		}

		input := &guardduty.InviteMembersInput{
			DetectorId: aws.String(detectorID),
			AccountIds: []*string{aws.String(accountID)},
			Message: aws.String(message),
		}

		imo, err := conn.InviteMembers(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		// Fix me
		if len(imo.UnprocessedAccounts) > 0 {
			continue
		}

		return fmt.Errorf("Expected GuardDuty Detector to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyInviteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		accountID, detectorID, message, err := decodeGuardDutyInviteID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.InviteMembersInput{
			DetectorId: aws.String(detectorID),
			AccountIds: []*string{aws.String(accountID)},
			Message: aws.String(message),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		imo, err := conn.InviteMembers(input)
		if err != nil {
			return err
		}

		if len(imo.UnprocessedAccounts) >0  {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccGuardDutyInviteConfig_basic(accountID, email, message string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id  = "%[2]s"
  detector_id = "${aws_guardduty_detector.test.id}"
  email       = "%[3]s"
}
`	, testAccGuardDutyDetectorConfig_basic1, accountID, email, message)
}