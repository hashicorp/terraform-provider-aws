// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccountAlias_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DataSource": {
			acctest.CtBasic: testAccAccountAliasDataSource_basic,
		},
		"Resource": {
			acctest.CtBasic: testAccAccountAlias_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccAccountAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_account_alias.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckAccountAlias(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_account_alias" {
				continue
			}

			params := &iam.ListAccountAliasesInput{}

			resp, err := conn.ListAccountAliases(ctx, params)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliases(ctx, params)

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

func testAccPreCheckAccountAlias(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

	input := &iam.CreateAccountAliasInput{
		AccountAlias: aws.String(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
	}
	_, err := conn.CreateAccountAlias(ctx, input)

	if tfawserr.ErrCodeEquals(err, "AccessDenied") {
		t.Skip("skipping acceptance testing: AccessDenied")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAccountAliasConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_account_alias" "test" {
  account_alias = %[1]q
}
`, rName)
}
