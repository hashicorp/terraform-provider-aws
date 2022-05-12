package kms_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSCiphertext_Resource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCiphertextConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"aws_kms_ciphertext.foo", "ciphertext_blob"),
				),
			},
		},
	})
}

func TestAccKMSCiphertext_Resource_validate(t *testing.T) {
	kmsSecretsDataSource := "data.aws_kms_secrets.foo"
	resourceName := "aws_kms_ciphertext.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCiphertextConfig_validate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

func TestAccKMSCiphertext_ResourceValidate_withContext(t *testing.T) {
	kmsSecretsDataSource := "data.aws_kms_secrets.foo"
	resourceName := "aws_kms_ciphertext.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCiphertextConfig_validate_withContext,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

const testAccResourceCiphertextConfig_basic = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-basic"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"
}
`

const testAccResourceCiphertextConfig_validate = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"
}

data "aws_kms_secrets" "foo" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.foo.ciphertext_blob
  }
}
`

const testAccResourceCiphertextConfig_validate_withContext = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate-with-context"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"

  context = {
    name = "value"
  }
}

data "aws_kms_secrets" "foo" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.foo.ciphertext_blob

    context = {
      name = "value"
    }
  }
}
`
