package macie_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie "github.com/hashicorp/terraform-provider-aws/internal/service/macie"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMacieMemberAccountAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "MACIE_MEMBER_ACCOUNT_ID"
	memberAccountID := os.Getenv(key)
	if memberAccountID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	resourceName := "aws_macie_member_account_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, macie.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberAccountAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAccountAssociationConfig_basic(memberAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAccountAssociationExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccMacieMemberAccountAssociation_self(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_macie_member_account_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, macie.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAccountAssociationConfig_self,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAccountAssociationExists(ctx, resourceName),
				),
			},
		},
	})
}

func testAccCheckMemberAccountAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie_member_account_association" {
				continue
			}

			_, err := tfmacie.FindMemberAccountByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Macie Classic Member Account Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMemberAccountAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Macie Classic Member Account Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MacieConn()

		_, err := tfmacie.FindMemberAccountByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccMemberAccountAssociationConfig_basic(accountID string) string {
	return fmt.Sprintf(`
resource "aws_macie_member_account_association" "test" {
  member_account_id = %[1]q
}
`, accountID)
}

const testAccMemberAccountAssociationConfig_self = `
data "aws_caller_identity" "current" {}

resource "aws_macie_member_account_association" "test" {
  member_account_id = data.aws_caller_identity.current.account_id
}
`
