// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_redshiftserverless_snapshot.test", names.AttrARN),
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

func TestAccRedshiftServerlessResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshiftserverless.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_resource_policy" {
				continue
			}
			_, err := tfredshiftserverless.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Resource Policy is not set")
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftServerlessClient(ctx)

		_, err := tfredshiftserverless.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "test" {
  provider = "awsalternate"
}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_snapshot" "test" {
  namespace_name = aws_redshiftserverless_workgroup.test.namespace_name
  snapshot_name  = %[1]q
}

resource "aws_redshiftserverless_resource_policy" "test" {
  resource_arn = aws_redshiftserverless_snapshot.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = [data.aws_caller_identity.test.account_id]
      }
      Action = [
        "redshift-serverless:RestoreFromSnapshot",
      ]
      Sid = ""
    }]
  })
}
`, rName))
}
