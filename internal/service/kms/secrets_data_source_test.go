package kms_test

import (
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
	var encryptedPayload string
	var key kms.KeyMetadata

	plaintext := "my-plaintext-string"
	resourceName := "aws_kms_key.test"

	// Run a resource test to setup our KMS key
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsDataSourceConfig_key,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					testAccSecretsEncryptDataSource(&key, plaintext, &encryptedPayload),
					// We need to dereference the encryptedPayload in a test Terraform configuration
					testAccSecretsDecryptDataSource(t, plaintext, &encryptedPayload),
				),
			},
		},
	})
}

func testAccSecretsEncryptDataSource(key *kms.KeyMetadata, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

		input := &kms.EncryptInput{
			KeyId:     key.Arn,
			Plaintext: []byte(plaintext),
			EncryptionContext: map[string]*string{
				"name": aws.String("value"),
			},
		}

		resp, err := conn.Encrypt(input)
		if err != nil {
			return fmt.Errorf("failed encrypting string: %s", err)
		}

		*encryptedPayload = base64.StdEncoding.EncodeToString(resp.CiphertextBlob)

		return nil
	}
}

func testAccSecretsDecryptDataSource(t *testing.T, plaintext string, encryptedPayload *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSourceName := "data.aws_kms_secrets.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { acctest.PreCheck(t) },
			ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
			ProviderFactories: acctest.ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccCheckSecretsSecretDataSource(*encryptedPayload),
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

func testAccCheckSecretsSecretDataSource(payload string) string {
	return testAccSecretsDataSourceConfig_key + fmt.Sprintf(`
data "aws_kms_secrets" "test" {
  secret {
    name    = "secret1"
    payload = %q

    context = {
      name = "value"
    }
  }
}
`, payload)
}
