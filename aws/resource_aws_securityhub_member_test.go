package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSecurityHubMember_basic(t *testing.T) {
	var member securityhub.Member

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		t.Skip("'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists("aws_securityhub_member.example", &member),
				),
			},
			{
				ResourceName:      "aws_securityhub_member.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
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
			return fmt.Errorf("API Gateway Resource still exists")
		}

		return nil
	}

	return nil
}

func testAccAWSSecurityHubMemberConfig(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "example" {
  name  = "%s"
  email = "%s"
}

resource "aws_securityhub_member" "example" {
  account_id = "${aws_organizations_account.example.id}"
  email      = "%s"
}
`, name, email, email)
}
