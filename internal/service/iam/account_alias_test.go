package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMAccountAlias_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DataSource": {
			"basic": testAccAccountAliasDataSource_basic,
		},
		"Resource": {
			"basic": testAccAccountAlias_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccAccountAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_account_alias.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountAliasExists(ctx, resourceName),
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

func testAccCheckAccountAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_account_alias" {
				continue
			}

			params := &iam.ListAccountAliasesInput{}

			resp, err := conn.ListAccountAliasesWithContext(ctx, params)

			if err != nil {
				return fmt.Errorf("error reading IAM Account Alias (%s): %w", rs.Primary.ID, err)
			}

			if resp == nil {
				return fmt.Errorf("error reading IAM Account Alias (%s): empty response", rs.Primary.ID)
			}

			if len(resp.AccountAliases) > 0 {
				return fmt.Errorf("Bad: Account alias still exists: %q", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckAccountAliasExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()
		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliasesWithContext(ctx, params)

		if err != nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): %w", rs.Primary.ID, err)
		}

		if resp == nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): empty response", rs.Primary.ID)
		}

		if len(resp.AccountAliases) == 0 {
			return fmt.Errorf("Bad: Account alias %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccountAliasConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_account_alias" "test" {
  account_alias = %[1]q
}
`, rName)
}
