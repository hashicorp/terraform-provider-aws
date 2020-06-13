---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_domain"
description: |-
  Provides a CodeArtifact Domain resource.
---

# Resource: aws_codeartifact_domain

Provides a CodeArtifact Domains Resource.

## Example Usage

```hcl
resource "aws_kms_key" "example" {
  description = "domain key"
}

resource "aws_codeartifact_domain" "example" {
  domain         = "example.com"
  encryption_key = "${aws_kms_key.example.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The name of the domain to create. All domain names in an AWS Region that are in the same AWS account must be unique. The domain name is used as the prefix in DNS hostnames. Do not use sensitive information in a domain name because it is publicly discoverable.
* `encryption_key` - (Required) The encryption key for the domain. This is used to encrypt content stored in a domain. An encryption key can be a key ID, a key Amazon Resource Name (ARN), a key alias, or a key alias ARN.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Name of Domain.
* `arn` - The ARN of Domain.
* `owner` - The AWS account ID that owns the domain.
* `repository_count` - The number of repositories in the domain.
* `created_time` - A timestamp that represents the date and time the domain was created.
* `asset_size_bytes` - The total size of all assets in the domain.

## Import

CodeArtifact Domain can be imported using the CodeArtifact Domain name, e.g.

```
$ terraform import aws_codeartifact_domain.example example.com
```
