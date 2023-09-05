---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_access_group"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Verified Access Group.
---

# Resource: aws_verifiedaccess_access_group

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Verified Access Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedaccess_access_group" "example" {
  verified_access_instance_id = ""
}
```

## Argument Reference

The following arguments are required:

* `verified_access_instance_id` - (Required) The id of the verified access instance this group is associated with.

The following arguments are optional:

* `description` - (Optional) Description of the verified access group.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `policy_docment` - (Optional) The policy document that is associated with this resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `creation_time` - Timestamp when the access group was created.
* `deletion_time` - Timestamp when the access group was deleted.
* `last_updated_time` - Timestamp when the access group was last updated.
* `owner` - AWS account number owning this resource.
* `verified_access_group_arn` - ARN of this verified acess group.
* `verified_access_group_id` - ID of this verified access group.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)
