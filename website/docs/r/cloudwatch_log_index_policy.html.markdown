---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_index_policy"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Index Policy.
---

# Resource: aws_cloudwatch_log_index_policy

Terraform resource for managing an AWS CloudWatch Logs Index Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_index_policy" "example" {
  log_group_name = aws_cloudwatch_log_group.example.name
  policy_document = jsonencode({
    Fields = ["eventName"]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `log_group_name` - (Required) Log group name to set the policy for.
* `policy_document` - (Required) JSON policy document. This is a JSON formatted string.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_index_policy.example
  identity = {
    log_group_name = "/aws/log/group/name"
  }
}

resource "aws_cloudwatch_log_index_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `log_group_name` (String) Name of the log group.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Index Policies using `log_group_name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_index_policy.example
  id = "/aws/log/group/name"
}
```

Using `terraform import`, import Index Policies using `log_group_name`. For example:

```console
% terraform import aws_cloudwatch_log_index_policy.example /aws/log/group/name
```
