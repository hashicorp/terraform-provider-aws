// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glacier_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glacier"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglacier "github.com/hashicorp/terraform-provider-aws/internal/service/glacier"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlacierVaultLock_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultLockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_complete(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(ctx, resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, names.AttrName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ignore_deletion_error"},
			},
		},
	})
}

func TestAccGlacierVaultLock_completeLock(t *testing.T) {
	ctx := acctest.Context(t)
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultLockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_complete(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(ctx, resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, names.AttrName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ignore_deletion_error"},
			},
		},
	})
}

func TestAccGlacierVaultLock_ignoreEquivalentPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultLockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_policyOrder(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(ctx, resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, names.AttrName),
				),
			},
			{
				Config:   testAccVaultLockConfig_policyNewOrder(rName, false),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVaultLockExists(ctx context.Context, n string, v *glacier.GetVaultLockOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glacier Vault Lock ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierClient(ctx)

		output, err := tfglacier.FindVaultLockByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVaultLockDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glacier_vault_lock" {
				continue
			}

			_, err := tfglacier.FindVaultLockByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glacier Vault Lock %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVaultLockConfig_complete(rName string, completeLock bool) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %q
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    # Allow for testing purposes
    actions   = ["glacier:DeleteArchive"]
    effect    = "Allow"
    resources = [aws_glacier_vault.test.arn]

    condition {
      test     = "NumericLessThanEquals"
      variable = "glacier:ArchiveAgeinDays"
      values   = ["0"]
    }

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
  }
}

resource "aws_glacier_vault_lock" "test" {
  complete_lock         = %t
  ignore_deletion_error = %t
  policy                = data.aws_iam_policy_document.test.json
  vault_name            = aws_glacier_vault.test.name
}
`, rName, completeLock, completeLock)
}

func testAccVaultLockConfig_policyOrder(rName string, completeLock bool) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_glacier_vault_lock" "test" {
  complete_lock         = %[2]t
  ignore_deletion_error = %[2]t
  vault_name            = aws_glacier_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Effect = "Allow"
      Action = [
        "glacier:InitiateMultipartUpload",
        "glacier:AbortMultipartUpload",
        "glacier:CompleteMultipartUpload",
        "glacier:DeleteArchive",
      ]
      Resource = aws_glacier_vault.test.arn
    }]
  })
}
`, rName, completeLock)
}

func testAccVaultLockConfig_policyNewOrder(rName string, completeLock bool) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_glacier_vault_lock" "test" {
  complete_lock         = %[2]t
  ignore_deletion_error = %[2]t
  vault_name            = aws_glacier_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Principal = {
        AWS = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      }
      Effect = "Allow"
      Action = [
        "glacier:InitiateMultipartUpload",
        "glacier:DeleteArchive",
        "glacier:CompleteMultipartUpload",
        "glacier:AbortMultipartUpload",
      ]
      Resource = [aws_glacier_vault.test.arn]
    }]
  })
}
`, rName, completeLock)
}
