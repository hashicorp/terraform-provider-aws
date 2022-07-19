package glacier_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccGlacierVaultLock_basic(t *testing.T) {
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glacier.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultLockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_complete(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", "false"),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, "name"),
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
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glacier.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultLockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_complete(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", "true"),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, "name"),
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
	var vaultLock1 glacier.GetVaultLockOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vaultResourceName := "aws_glacier_vault.test"
	resourceName := "aws_glacier_vault_lock.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glacier.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultLockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultLockConfig_policyOrder(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockExists(resourceName, &vaultLock1),
					resource.TestCheckResourceAttr(resourceName, "complete_lock", "false"),
					resource.TestCheckResourceAttr(resourceName, "ignore_deletion_error", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_name", vaultResourceName, "name"),
				),
			},
			{
				Config:   testAccVaultLockConfig_policyNewOrder(rName, false),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVaultLockExists(resourceName string, getVaultLockOutput *glacier.GetVaultLockOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierConn

		input := &glacier.GetVaultLockInput{
			VaultName: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetVaultLock(input)

		if err != nil {
			return fmt.Errorf("error reading Glacier Vault Lock (%s): %s", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("error reading Glacier Vault Lock (%s): empty response", rs.Primary.ID)
		}

		*getVaultLockOutput = *output

		return nil
	}
}

func testAccCheckVaultLockDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glacier_vault_lock" {
			continue
		}

		input := &glacier.GetVaultLockInput{
			VaultName: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetVaultLock(input)

		if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Glacier Vault Lock (%s): %s", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Glacier Vault Lock (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
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
