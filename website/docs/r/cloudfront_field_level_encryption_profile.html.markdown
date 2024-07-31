---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_field_level_encryption_profile"
description: |-
  Provides a CloudFront Field-level Encryption Profile resource.
---

# Resource: aws_cloudfront_field_level_encryption_profile

Provides a CloudFront Field-level Encryption Profile resource.

## Example Usage

```terraform
resource "aws_cloudfront_public_key" "example" {
  comment     = "test public key"
  encoded_key = file("public_key.pem")
  name        = "test_key"
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "test comment"
  name    = "test profile"

  encryption_entities {
    items {
      public_key_id = aws_cloudfront_public_key.example.id
      provider_id   = "test provider"

      field_patterns {
        items = ["DateOfBirth"]
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the Field Level Encryption Profile.
* `comment` - (Optional) An optional comment about the Field Level Encryption Profile.
* `encryption_entities` - (Required) The [encryption entities](#encryption-entities) config block for field-level encryption profiles that contains an attribute `items` which includes the encryption key and field pattern specifications.

### Encryption Entities

* `public_key_id` - (Required) The public key associated with a set of field-level encryption patterns, to be used when encrypting the fields that match the patterns.
* `provider_id` - (Required) The provider associated with the public key being used for encryption.
* `field_patterns` - (Required) Object that contains an attribute `items` that contains the list of field patterns in a field-level encryption content type profile specify the fields that you want to be encrypted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `caller_reference` - Internal value used by CloudFront to allow future updates to the Field Level Encryption Profile.
* `etag` - The current version of the Field Level Encryption Profile. For example: `E2QWRUHAPOMQZL`.
* `id` - The identifier for the Field Level Encryption Profile. For example: `K3D5EWEUDCCXON`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudfront Field Level Encryption Profile using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_field_level_encryption_profile.profile
  id = "K3D5EWEUDCCXON"
}
```

Using `terraform import`, import Cloudfront Field Level Encryption Profile using the `id`. For example:

```console
% terraform import aws_cloudfront_field_level_encryption_profile.profile K3D5EWEUDCCXON
```
