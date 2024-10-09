// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSSecretsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var encryptedPayload string
	var key awstypes.KeyMetadata

	plaintext := "my-plaintext-string"
	resourceName := "aws_kms_key.test"

	// Run a resource test to setup our KMS key
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_key,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					testAccSecretsDataSourceEncrypt(ctx, &key, plaintext, &encryptedPayload),
					// We need to dereference the encryptedPayload in a test Terraform configuration
					testAccSecretsDataSourceDecrypt(ctx, t, plaintext, &encryptedPayload),
				),
			},
		},
	})
}

func TestAccKMSSecretsDataSource_asymmetric(t *testing.T) {
	ctx := acctest.Context(t)
	var encryptedPayload string
	var key awstypes.KeyMetadata

	plaintext := "my-plaintext-string"
	resourceName := "aws_kms_key.test"

	// Run a resource test to setup our KMS key
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_asymmetricKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					testAccSecretsDataSourceEncryptAsymmetric(ctx, &key, plaintext, &encryptedPayload),
					// We need to dereference the encryptedPayload in a test Terraform configuration
					testAccSecretsDataSourceDecryptAsymmetric(ctx, t, &key, plaintext, &encryptedPayload),
				),
			},
		},
	})
}

func testAccSecretsDataSourceEncrypt(ctx context.Context, key *awstypes.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

		input := &kms.EncryptInput{
			KeyId:     key.Arn,
			Plaintext: []byte(plaintext),
			EncryptionContext: map[string]string{
				names.AttrName: names.AttrValue,
			},
		}

		output, err := conn.Encrypt(ctx, input)

		if err != nil {
			return err
		}

		*encryptedPayload = base64.StdEncoding.EncodeToString(output.CiphertextBlob)

		return nil
	}
}

func testAccSecretsDataSourceEncryptAsymmetric(ctx context.Context, key *awstypes.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

		input := &kms.EncryptInput{
			KeyId:               key.Arn,
			Plaintext:           []byte(plaintext),
			EncryptionAlgorithm: awstypes.EncryptionAlgorithmSpecRsaesOaepSha1,
		}

		output, err := conn.Encrypt(ctx, input)

		if err != nil {
			return err
		}

		*encryptedPayload = base64.StdEncoding.EncodeToString(output.CiphertextBlob)

		return nil
	}
}

func testAccSecretsDataSourceDecrypt(ctx context.Context, t *testing.T, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSourceName := "data.aws_kms_secrets.test"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(ctx, t) },
			ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecretsDataSourceConfig_secret(*encryptedPayload),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.%", acctest.Ct1),
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.secret1", plaintext),
					),
				},
			},
		})

		return nil
	}
}

func testAccSecretsDataSourceDecryptAsymmetric(ctx context.Context, t *testing.T, key *awstypes.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSourceName := "data.aws_kms_secrets.test"
		keyid := key.Arn

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(ctx, t) },
			ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecretsDataSourceConfig_asymmetricSecret(*encryptedPayload, *keyid),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "plaintext.%", acctest.Ct1),
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
