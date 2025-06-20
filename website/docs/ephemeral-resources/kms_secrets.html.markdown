---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_secrets"
description: |-
    Decrypt multiple secrets from data encrypted with the AWS KMS service
---

# Ephemeral: aws_kms_secrets

Decrypt multiple secrets from data encrypted with the AWS KMS service.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

If you do not already have a `CiphertextBlob` from encrypting a KMS secret, you can use the below commands to obtain one using the [AWS CLI kms encrypt](https://docs.aws.amazon.com/cli/latest/reference/kms/encrypt.html) command. This requires you to have your AWS CLI setup correctly and replace the `--key-id` with your own. Alternatively you can use `--plaintext 'master-password'` (CLIv1) or `--plaintext fileb://<(echo -n 'master-password')` (CLIv2) instead of reading from a file.

-> If you have a newline character at the end of your file, it will be decrypted with this newline character intact. For most use cases this is undesirable and leads to incorrect passwords or invalid values, as well as possible changes in the plan. Be sure to use `echo -n` if necessary.
-> If you are using asymmetric keys ensure you are using the right encryption algorithm when you encrypt and decrypt else you will get IncorrectKeyException during the decrypt phase.

That encrypted output can now be inserted into Terraform configurations without exposing the plaintext secret directly.

```terraform
ephemeral "aws_kms_secrets" "example" {
  secret {
    # ... potentially other configuration ...
    name    = "master_password"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="

    context = {
      foo = "bar"
    }
  }

  secret {
    # ... potentially other configuration ...
    name    = "master_username"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = data.aws_kms_secrets.example.plaintext["master_password"]
  master_username = data.aws_kms_secrets.example.plaintext["master_username"]
}

ephemeral "aws_kms_secrets" "example" {
  secret {
    # ... potentially other configuration ...
    name    = "app_specific_secret"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="
    # ... Use same algorithm used to Encrypt the payload ...
    encryption_algorithm = "RSAES_OAEP_SHA_256"
    key_id               = "ab123456-c012-4567-890a-deadbeef123"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `secret` - (Required) One or more encrypted payload definitions from the KMS service. See the Secret Definitions below.

### Secret Definitions

Each `secret` supports the following arguments:

* `name` - (Required) Name to export this secret under in the attributes.
* `payload` - (Required) Base64 encoded payload, as returned from a KMS encrypt operation.
* `context` - (Optional) An optional mapping that makes up the Encryption Context for the secret.
* `grant_tokens` (Optional) An optional list of Grant Tokens for the secret.
* `encryption_algorithm` - (Optional) The encryption algorithm that will be used to decrypt the ciphertext. This parameter is required only when the ciphertext was encrypted under an asymmetric KMS key. Valid Values: SYMMETRIC_DEFAULT | RSAES_OAEP_SHA_1 | RSAES_OAEP_SHA_256 | SM2PKE
* `key_id` (Optional) Specifies the KMS key that AWS KMS uses to decrypt the ciphertext. This parameter is required only when the ciphertext was encrypted under an asymmetric KMS key.

For more information on `context` and `grant_tokens` see the [KMS
Concepts](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `plaintext` - Map containing each `secret` `name` as the key with its decrypted plaintext value
