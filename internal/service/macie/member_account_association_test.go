package macie_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMacieMemberAccountAssociation_basic(t *testing.T) {
	key := "MACIE_MEMBER_ACCOUNT_ID"
	memberAcctId := os.Getenv(key)
	if memberAcctId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, macie.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMemberAccountAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAccountAssociationConfig_basic(memberAcctId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAccountAssociationExists("aws_macie_member_account_association.test"),
				),
			},
		},
	})
}

func TestAccMacieMemberAccountAssociation_self(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, macie.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// master account associated with Macie it can't be disassociated.
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAccountAssociationConfig_self,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAccountAssociationExists("aws_macie_member_account_association.test"),
				),
			},
		},
	})
}

func testAccCheckMemberAccountAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie_member_account_association" {
			continue
		}

		req := &macie.ListMemberAccountsInput{}

		dissociated := true
		err := conn.ListMemberAccountsPages(req, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
			for _, v := range page.MemberAccounts {
				if aws.StringValue(v.AccountId) == rs.Primary.Attributes["member_account_id"] {
					dissociated = false
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !dissociated {
			return fmt.Errorf("Member account %s is not dissociated from Macie", rs.Primary.Attributes["member_account_id"])
		}
	}
	return nil
}

func testAccCheckMemberAccountAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		req := &macie.ListMemberAccountsInput{}

		exists := false
		err := conn.ListMemberAccountsPages(req, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
			for _, v := range page.MemberAccounts {
				if aws.StringValue(v.AccountId) == rs.Primary.Attributes["member_account_id"] {
					exists = true
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Member account %s is not associated with Macie", rs.Primary.Attributes["member_account_id"])
		}

		return nil
	}
}

func testAccMemberAccountAssociationConfig_basic(memberAcctId string) string {
	return fmt.Sprintf(`
resource "aws_macie_member_account_association" "test" {
  member_account_id = "%s"
}
`, memberAcctId)
}

const testAccMemberAccountAssociationConfig_self = `
data "aws_caller_identity" "current" {}

resource "aws_macie_member_account_association" "test" {
  member_account_id = data.aws_caller_identity.current.account_id
}
`
