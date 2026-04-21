// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccountAlias_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DataSource": {
			acctest.CtBasic: testAccAccountAliasDataSource_basic,
		},
		"Resource": {
			acctest.CtBasic:      testAccAccountAlias_basic,
			acctest.CtDisappears: testAccAccountAlias_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccAccountAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_account_alias.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckAccountAlias(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountAliasExists(ctx, t, resourceName),
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

func testAccAccountAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_account_alias.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckAccountAlias(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountAliasExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceAccountAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_account_alias" {
				continue
			}

			var input iam.ListAccountAliasesInput
			_, err := tfiam.FindAccountAlias(ctx, conn, &input)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Server Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccountAliasExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		var input iam.ListAccountAliasesInput
		_, err := tfiam.FindAccountAlias(ctx, conn, &input)

		return err
	}
}

func testAccPreCheckAccountAlias(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

	input := &iam.CreateAccountAliasInput{
		AccountAlias: aws.String(acctest.RandomWithPrefix(t, acctest.ResourcePrefix)),
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
