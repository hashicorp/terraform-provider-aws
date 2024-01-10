---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_group"
description: |-
  Terraform resource for managing a Verified Access Group.
---

# Resource: aws_verifiedaccess_group

Terraform resource for managing a Verified Access Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedaccess_group" "example" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

## Argument Reference

The following arguments are required:

* `verifiedaccess_instance_id` - (Required) The id of the verified access instance this group is associated with.

The following arguments are optional:

* `description` - (Optional) Description of the verified access group.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `policy_document` - (Optional) The policy document that is associated with this resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - Timestamp when the access group was created.
* `deletion_time` - Timestamp when the access group was deleted.
* `last_updated_time` - Timestamp when the access group was last updated.
* `owner` - AWS account number owning this resource.
* `verifiedaccess_group_arn` - ARN of this verified acess group.
* `verifiedaccess_group_id` - ID of this verified access group.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)
