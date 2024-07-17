// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSCiphertext_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

const testAccCiphertextConfig_basic = `
resource "aws_kms_key" "test" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-basic"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"
}
`

const testAccCiphertextConfig_validate = `
resource "aws_kms_key" "test" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate"
  is_enabled  = true
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
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate-with-context"
  is_enabled  = true
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
