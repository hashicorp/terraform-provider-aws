---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_ip_group"
description: |-
  Provides an IP access control group in AWS WorkSpaces Service.
---

# Resource: aws_workspaces_ip_group

Provides an IP access control group in AWS WorkSpaces Service

## Example Usage

```terraform
resource "aws_workspaces_ip_group" "contractors" {
  name        = "Contractors"
  description = "Contractors IP access control group"
  rules {
    source      = "150.24.14.0/24"
    description = "NY"
  }
  rules {
    source      = "125.191.14.85/32"
    description = "LA"
  }
  rules {
    source      = "44.98.100.0/24"
    description = "STL"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `description` - (Optional) Description of the IP group.
* `name` - (Required) Name of the IP group.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rules` - (Optional) One or more pairs specifying the IP group rule (in CIDR format) from which web requests originate.
* `tags` - (Optional) Map of tags assigned to the WorkSpaces IP group. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rules` Block

* `description` - (Optional) Description of the IP rule.
* `source` - (Required) IP address range, in CIDR notation, e.g., `10.0.0.0/16`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - IP group identifier.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces IP groups using their GroupID. For example:

```terraform
import {
  to = aws_workspaces_ip_group.example
  id = "wsipg-488lrtl3k"
}
```

Using `terraform import`, import WorkSpaces IP groups using their GroupID. For example:

```console
% terraform import aws_workspaces_ip_group.example wsipg-488lrtl3k
```
