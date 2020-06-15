---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_domain_permissions"
description: |-
  Provides a CodeArtifact Domain Permissions resource.
---

# Resource: aws_codeartifact_domain_permissions

Provides a CodeArtifact Domains Permissions Resource.

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

* `domain` - (Required) The name of the domain on which to set the resource policy.
* `encryption_key` - (Required) A valid displayable JSON Aspen policy string to be set as the access control resource policy on the provided domain.
* `policy_revision` - (Optional) The current revision of the resource policy to be set. This revision is used for optimistic locking, which prevents others from overwriting your changes to the domain's resource policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Name of Domain.
* `resource_arn` - The ARN of the resource associated with the resource policy.

## Import

CodeArtifact Domain Permissions can be imported using the CodeArtifact Domain name, e.g.

```
$ terraform import aws_codeartifact_domain_permissions.example example.com
```
