// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSCiphertext_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"aws_kms_ciphertext.test", "ciphertext_blob"),
				),
			},
		},
	})
}

func TestAccKMSCiphertext_validate(t *testing.T) {
	ctx := acctest.Context(t)
	kmsSecretsDataSource := "data.aws_kms_secrets.test"
	resourceName := "aws_kms_ciphertext.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextConfig_validate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

func TestAccKMSCiphertext_Validate_withContext(t *testing.T) {
	ctx := acctest.Context(t)
	kmsSecretsDataSource := "data.aws_kms_secrets.test"
	resourceName := "aws_kms_ciphertext.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextConfig_validateContext,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

func TestAccKMSCiphertext_plaintextWO(t *testing.T) {
	ctx := acctest.Context(t)
	kmsSecretsDataSource := "data.aws_kms_secrets.test"
	resourceName := "aws_kms_ciphertext.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextConfig_plaintextWO("Secret1", "1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectSensitiveValue(resourceName, tfjsonpath.New("plaintext_wo")),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttr(kmsSecretsDataSource, "plaintext.plaintext", "Secret1"),
				),
			},
			{
				Config: testAccCiphertextConfig_plaintextWO("Secret2", "2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
						plancheck.ExpectSensitiveValue(resourceName, tfjsonpath.New("plaintext_wo")),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttr(kmsSecretsDataSource, "plaintext.plaintext", "Secret2"),
				),
			},
		},
	})
}

const testAccCiphertextConfig_basic = `
resource "aws_kms_key" "test" {
  description             = "tf-test-acc-data-source-aws-kms-ciphertext-basic"
  is_enabled              = true
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"
}
`

const testAccCiphertextConfig_validate = `
resource "aws_kms_key" "test" {
  description             = "tf-test-acc-data-source-aws-kms-ciphertext-validate"
  is_enabled              = true
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"
}

data "aws_kms_secrets" "test" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.test.ciphertext_blob
  }
}
`

const testAccCiphertextConfig_validateContext = `
resource "aws_kms_key" "test" {
  description             = "tf-test-acc-data-source-aws-kms-ciphertext-validate-with-context"
  is_enabled              = true
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"

  context = {
    name = "value"
  }
}

data "aws_kms_secrets" "test" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.test.ciphertext_blob

    context = {
      name = "value"
    }
  }
}
`

func testAccCiphertextConfig_plaintextWO(plaintextWO, plaintextWOVersion string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "tf-test-acc-data-source-aws-kms-ciphertext-plaintext-wo"
  is_enabled              = true
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext_wo         = %[1]q
  plaintext_wo_version = %[2]q
}

data "aws_kms_secrets" "test" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.test.ciphertext_blob
  }
}

`, plaintextWO, plaintextWOVersion)
}
