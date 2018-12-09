package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSecurityHubMember_basic(t *testing.T) {
	var member securityhub.Member
	resourceName := "aws_securityhub_member.example"
	accountId, email := testAccAWSSecurityHubMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig_basic(accountId, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
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

func TestAccAWSSecurityHubMember_invite_create(t *testing.T) {
	var member securityhub.Member
	resourceName := "aws_securityhub_member.example"
	accountId, email := testAccAWSSecurityHubMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig_invite(accountId, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Invited"),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
				),
			},
			// Disassociate member
			{
				Config: testAccAWSSecurityHubMemberConfig_invite(accountId, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
					// These may fail: https://github.com/aws/aws-sdk-go/issues/2332#issuecomment-445535357
					resource.TestCheckResourceAttr(resourceName, "member_status", "Removed"),
					resource.TestCheckResourceAttr(resourceName, "invite", "false"),
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

func TestAccAWSSecurityHubMember_invite_update(t *testing.T) {
	var member securityhub.Member
	resourceName := "aws_securityhub_member.example"
	accountId, email := testAccAWSSecurityHubMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig_invite(accountId, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Created"),
					resource.TestCheckResourceAttr(resourceName, "invite", "false"),
				),
			},
			// Invite member
			{
				Config: testAccAWSSecurityHubMemberConfig_invite(accountId, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Invited"),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
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

func testAccAWSSecurityHubMemberFromEnv(t *testing.T) (string, string) {
	accountId := os.Getenv("AWS_ALTERNATE_ACCOUNT_ID")
	if accountId == "" {
		accountId = "111111111111"
	}

	email := os.Getenv("AWS_ALTERNATE_EMAIL")
	if email == "" {
		email = "example@example.com"
	}

	return accountId, email
}

func testAccCheckAWSSecurityHubMemberExists(n string, member *securityhub.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		resp, err := conn.GetMembers(&securityhub.GetMembersInput{
			AccountIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(resp.Members) == 0 {
			return fmt.Errorf("Security Hub member %s not found", rs.Primary.ID)
		}

		member = resp.Members[0]

		return nil
	}
}

func testAccCheckAWSSecurityHubMemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_member" {
			continue
		}

		resp, err := conn.GetMembers(&securityhub.GetMembersInput{
			AccountIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if len(resp.Members) != 0 {
			return fmt.Errorf("Security Hub member still exists")
		}

		return nil
	}

	return nil
}

func testAccAWSSecurityHubMemberConfig_basic(accountId, email string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = ["aws_securityhub_account.example"]
  account_id = "%s"
  email      = "%s"
}
`, accountId, email)
}

func testAccAWSSecurityHubMemberConfig_invite(accountId, email string, invite bool) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = ["aws_securityhub_account.example"]
  account_id = "%s"
  email      = "%s"
  invite     = %t
}
`, accountId, email, invite)
}
