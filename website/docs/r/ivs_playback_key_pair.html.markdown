---
subcategory: "IVS (Interactive Video)"
layout: "aws"
page_title: "AWS: aws_ivs_playback_key_pair"
description: |-
  Terraform resource for managing an AWS IVS (Interactive Video) Playback Key Pair.
---

# Resource: aws_ivs_playback_key_pair

Terraform resource for managing an AWS IVS (Interactive Video) Playback Key Pair.

## Example Usage

### Basic Usage

```terraform
resource "aws_ivs_playback_key_pair" "example" {
  # public-key.pem is a file that contains an ECDSA public key in PEM format.
  public_key = file("./public-key.pem")
}
```

## Argument Reference

The following arguments are required:

* `public_key` - (Required) Public portion of a customer-generated key pair. Must be an ECDSA public key in PEM format.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) Playback Key Pair name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Playback Key Pair.
* `fingerprint` - Key-pair identifier.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IVS (Interactive Video) Playback Key Pair using the ARN. For example:

```terraform
import {
  to = aws_ivs_playback_key_pair.example
  id = "arn:aws:ivs:us-west-2:326937407773:playback-key/KDJRJNQhiQzA"
}
```

Using `terraform import`, import IVS (Interactive Video) Playback Key Pair using the ARN. For example:

```console
% terraform import aws_ivs_playback_key_pair.example arn:aws:ivs:us-west-2:326937407773:playback-key/KDJRJNQhiQzA
```
