---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_store"
description: |-
  This is a Terraform resource for managing an AWS Verified Permissions Policy Store.
---

# Resource: aws_verifiedpermissions_policy_store

This is a Terraform resource for managing an AWS Verified Permissions Policy Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_policy_store" "example" {
  validation_settings {
    mode = "STRICT"
  }
}
```

## Argument Reference

The following arguments are required:

* `validation_settings` - (Required) Validation settings for the policy store.
    * `mode` - (Required) The mode for the validation settings. Valid values: `OFF`, `STRICT`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) A description of the Policy Store.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_store_id` - The ID of the Policy Store.
* `arn` - The ARN of the Policy Store.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Permissions Policy Store using the `policy_store_id`. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy_store.example
  id = "DxQg2j8xvXJQ1tQCYNWj9T"
}
```

Using `terraform import`, import Verified Permissions Policy Store using the `policy_store_id`. For example:

```console
 % terraform import aws_verifiedpermissions_policy_store.example DxQg2j8xvXJQ1tQCYNWj9T
```
