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

* `name` - (Optional) Playback Key Pair name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Playback Key Pair.
* `fingerprint` - Key-pair identifier.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

IVS (Interactive Video) Playback Key Pair can be imported using the ARN, e.g.,

```
$ terraform import aws_ivs_playback_key_pair.example arn:aws:ivs:us-west-2:326937407773:playback-key/KDJRJNQhiQzA
```
