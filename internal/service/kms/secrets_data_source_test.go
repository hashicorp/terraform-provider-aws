package kms_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccKMSSecretsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var encryptedPayload string
	var key kms.KeyMetadata

	plaintext := "my-plaintext-string"
	resourceName := "aws_kms_key.test"

	// Run a resource test to setup our KMS key
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_key,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					testAccSecretsEncryptDataSource(ctx, &key, plaintext, &encryptedPayload),
					// We need to dereference the encryptedPayload in a test Terraform configuration
					testAccSecretsDecryptDataSource(t, plaintext, &encryptedPayload),
				),
			},
		},
	})
}

func TestAccKMSSecretsDataSource_asymmetric(t *testing.T) {
	ctx := acctest.Context(t)
	var encryptedPayload string
	var key kms.KeyMetadata

	plaintext := "my-plaintext-string"
	resourceName := "aws_kms_key.test"

	// Run a resource test to setup our KMS key
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_asymmetricKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					testAccSecretsEncryptDataSourceAsymmetric(ctx, &key, plaintext, &encryptedPayload),
					// We need to dereference the encryptedPayload in a test Terraform configuration
					testAccSecretsDecryptDataSourceAsym(t, &key, plaintext, &encryptedPayload),
				),
			},
		},
	})
}

func testAccSecretsEncryptDataSource(ctx context.Context, key *kms.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		input := &kms.EncryptInput{
			KeyId:     key.Arn,
			Plaintext: []byte(plaintext),
			EncryptionContext: map[string]*string{
				"name": aws.String("value"),
			},
		}

		output, err := conn.EncryptWithContext(ctx, input)

		if err != nil {
			return err
		}

		*encryptedPayload = base64.StdEncoding.EncodeToString(output.CiphertextBlob)

		return nil
	}
}

func testAccSecretsEncryptDataSourceAsymmetric(ctx context.Context, key *kms.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		input := &kms.EncryptInput{
			KeyId:               key.Arn,
			Plaintext:           []byte(plaintext),
			EncryptionAlgorithm: aws.String("RSAES_OAEP_SHA_1"),
		}

		output, err := conn.EncryptWithContext(ctx, input)

		if err != nil {
			return err
		}

		*encryptedPayload = base64.StdEncoding.EncodeToString(output.CiphertextBlob)

		return nil
	}
}

func testAccSecretsDecryptDataSource(t *testing.T, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSourceName := "data.aws_kms_secrets.test"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecretsDataSourceConfig_secret(*encryptedPayload),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.%", "1"),
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.secret1", plaintext),
					),
				},
			},
		})

		return nil
	}
}

func testAccSecretsDecryptDataSourceAsym(t *testing.T, key *kms.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSourceName := "data.aws_kms_secrets.test"
		keyid := key.Arn

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecretsDataSourceConfig_asymmetricSecret(*encryptedPayload, *keyid),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.%", "1"),
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.secret1", plaintext),
					),
				},
			},
		})

		return nil
	}
}

const testAccSecretsDataSourceConfig_key = `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Testing the Terraform AWS KMS Secrets data_source"
}
`

func testAccSecretsDataSourceConfig_secret(payload string) string {
	return acctest.ConfigCompose(testAccSecretsDataSourceConfig_key, fmt.Sprintf(`
data "aws_kms_secrets" "test" {
  secret {
    name    = "secret1"
    payload = %[1]q

    context = {
      name = "value"
    }
  }
}
`, payload))
}

const testAccSecretsDataSourceConfig_asymmetricKey = `
resource "aws_kms_key" "test" {
  deletion_window_in_days  = 7
  description              = "Testing the Terraform AWS KMS Secrets data_source"
  customer_master_key_spec = "RSA_2048"
}
`

func testAccSecretsDataSourceConfig_asymmetricSecret(payload string, keyid string) string {
	return acctest.ConfigCompose(testAccSecretsDataSourceConfig_asymmetricKey, fmt.Sprintf(`
data "aws_kms_secrets" "test" {
  secret {
    name                 = "secret1"
    payload              = %[1]q
    encryption_algorithm = "RSAES_OAEP_SHA_1"
    key_id               = %[2]q
  }
}
`, payload, keyid))
}
