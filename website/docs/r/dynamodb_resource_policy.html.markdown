---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_resource_policy"
description: |-
  Terraform resource for managing an AWS DynamoDB Resource Policy.
---

# Resource: aws_dynamodb_resource_policy

Terraform resource for managing an AWS DynamoDB Resource Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_dynamodb_resource_policy" "example" {
  resource_arn = aws_dynamodb_table.example.arn
  policy       = data.aws_iam_policy_document.test.json
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the DynamoDB resource to which the policy will be attached. The resources you can specify include tables and streams. You can control index permissions using the base table's policy. To specify the same permission level for your table and its indexes, you can provide both the table and index Amazon Resource Name (ARN)s in the Resource field of a given Statement in your policy document. Alternatively, to specify different permissions for your table, indexes, or both, you can define multiple Statement fields in your policy document.

* `policy` - (Required) n Amazon Web Services resource-based policy document in JSON format. The maximum size supported for a resource-based policy document is 20 KB. DynamoDB counts whitespaces when calculating the size of a policy against this limit. For a full list of all considerations that you should keep in mind while attaching a resource-based policy, see Resource-based policy considerations.

The following arguments are optional:

* `confirm_remove_self_resource_access` - (Optional) Set this parameter to true to confirm that you want to remove your permissions to change the policy of this resource in the future.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `revision_id` -  A unique string that represents the revision ID of the policy. If you are comparing revision IDs, make sure to always use string comparison logic.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB Resource Policy using the `example_id_arg`. For example:

```terraform
import {
  to = aws_dynamodb_resource_policy.example
  id = "arn:aws:dynamodb:us-east-1:1234567890:table/my-table"
}
```

Using `terraform import`, import DynamoDB Resource Policy using the `example_id_arg`. For example:

```console
% terraform import aws_dynamodb_resource_policy.example arn:aws:dynamodb:us-east-1:1234567890:table/my-table
```
