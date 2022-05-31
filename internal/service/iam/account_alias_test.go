package iam_test

import (
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
	testCases := map[string]map[string]func(t *testing.T){
		"DataSource": {
			"basic": testAccAccountAliasDataSource_basic,
		},
		"Resource": {
			"basic": testAccAccountAlias_basic,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAccountAlias_basic(t *testing.T) {
	resourceName := "aws_iam_account_alias.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAliasConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountAliasExists(resourceName),
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

func testAccCheckAccountAliasDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_account_alias" {
			continue
		}

		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliases(params)

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

func testAccCheckAccountAliasExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliases(params)

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

func testAccAccountAliasConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_account_alias" "test" {
  account_alias = %[1]q
}
`, rName)
}
