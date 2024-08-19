// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSCiphertextDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.aws_kms_ciphertext.test", "ciphertext_blob"),
				),
			},
		},
	})
}

func TestAccKMSCiphertextDataSource_validate(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextDataSourceConfig_validate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.aws_kms_ciphertext.test", "ciphertext_blob"),
				),
			},
		},
	})
}

func TestAccKMSCiphertextDataSource_Validate_withContext(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCiphertextDataSourceConfig_validateContext,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.aws_kms_ciphertext.test", "ciphertext_blob"),
				),
			},
		},
	})
}

const testAccCiphertextDataSourceConfig_basic = `
resource "aws_kms_key" "test" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-basic"
  is_enabled  = true
}

data "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"
}
`

const testAccCiphertextDataSourceConfig_validate = `
resource "aws_kms_key" "test" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate"
  is_enabled  = true
}

data "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"
}
`

const testAccCiphertextDataSourceConfig_validateContext = `
resource "aws_kms_key" "test" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate-with-context"
  is_enabled  = true
}

data "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  plaintext = "Super secret data"

  context = {
    name = "value"
  }
}
`
