---
subcategory: "Chime SDK Voice"
layout: "aws"
page_title: "AWS: aws_chimesdkvoice_voice_profile_domain"
description: |-
  Terraform resource for managing an AWS Chime SDK Voice Profile Domain.
---

# Resource: aws_chimesdkvoice_voice_profile_domain

Terraform resource for managing an AWS Chime SDK Voice Profile Domain.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "test" {
  description             = "KMS Key for Voice Profile Domain"
  deletion_window_in_days = 7
}

resource "aws_chimesdkvoice_voice_profile_domain" "test" {
  name = "ExampleVoiceProfileDomain"
  server_side_encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
  description = "My Voice Profile Domain"
  tags = {
    key1 = "value1"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of Voice Profile Domain.
* `server_side_encryption_configuration` - (Required) Configuration for server side encryption.
    * `kms_key_arn` - (Required) ARN for KMS Key.

The following arguments are optional:

* `description` - (Optional) Description of Voice Profile Domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Voice Profile Domain.
* `id` - ID of the Voice Profile Domain.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30s`)
* `update` - (Default `30s`)
* `delete` - (Default `30s`)
