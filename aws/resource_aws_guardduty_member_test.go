package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strconv"
	"github.com/hashicorp/terraform/helper/acctest"
	"regexp"
)

func testAccAwsGuardDutyMember_basic(t *testing.T) {
	resourceName := "aws_guardduty_member.test"
	accountID := "111111111111"
	email := "required@example.com"
	invitationMessage := "inviting"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_basic1(accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "invite", strconv.FormatBool(true)),
				),
			},
			{
				Config: testAccGuardDutyMemberConfig_basic2(accountID, email, invitationMessage, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "invite", strconv.FormatBool(false)),
					resource.TestCheckResourceAttr(resourceName, "invitation_message", invitationMessage),
				),
			},
		},
	})
}

func testAccAwsGuardDutyMember_import(t *testing.T) {
	resourceName := "aws_guardduty_member.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_basic1("111111111111", "required@example.com"),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsGuardDutyMemberInvitation_timeout(t *testing.T) {

	rEmail := acctest.RandString(10) + "@" + acctest.RandString(5) + ".com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_basic2("111111111111", rEmail, "test", true),
				ExpectError: regexp.MustCompile("Expected member to be invited but was in state: EmailVerificationFailed"),
			},
		},
	})
}

func testAccCheckAwsGuardDutyMemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_member" {
			continue
		}

		accountID, detectorID, err := decodeGuardDutyMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetMembersInput{
			AccountIds: []*string{aws.String(accountID)},
			DetectorId: aws.String(detectorID),
		}

		gmo, err := conn.GetMembers(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
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

func testAccCheckAwsGuardDutyMemberExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		accountID, detectorID, err := decodeGuardDutyMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetMembersInput{
			AccountIds: []*string{aws.String(accountID)},
			DetectorId: aws.String(detectorID),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		gmo, err := conn.GetMembers(input)
		if err != nil {
			return err
		}

		if len(gmo.Members) < 1 {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccGuardDutyMemberConfig_basic1(accountID, email string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id  = "%[2]s"
  detector_id = "${aws_guardduty_detector.test.id}"
  email       = "%[3]s"
}
`, testAccGuardDutyDetectorConfig_basic1, accountID, email)
}

func testAccGuardDutyMemberConfig_basic2(accountID, email, invitationMessage string, invite bool) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id  = "%[2]s"
  detector_id = "${aws_guardduty_detector.test.id}"
  email       = "%[3]s"
  invite      = "%[4]s"
  invite      = "%[5]s"
}
`, testAccGuardDutyDetectorConfig_basic1, accountID, email, strconv.FormatBool(invite), invitationMessage)
}
