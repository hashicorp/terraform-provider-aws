---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_stream"
description: |-
  Provides a CloudWatch Log Stream resource.
---

# Resource: aws_cloudwatch_log_stream

Provides a CloudWatch Log Stream resource.

## Example Usage

```terraform
resource "aws_cloudwatch_log_group" "yada" {
  name = "Yada"
}

resource "aws_cloudwatch_log_stream" "foo" {
  name           = "SampleLogStream1234"
  log_group_name = aws_cloudwatch_log_group.yada.name
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the log stream. Must not be longer than 512 characters and must not contain `:`
* `log_group_name` - (Required) The name of the log group under which the log stream is to be created.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) specifying the log stream.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_stream.example
  identity = {
    log_group_name = "example-group"
    name           = "example-stream"
  }
}

resource "aws_cloudwatch_log_stream" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `log_group_name` (String) Name of the log group.
* `name` (String) Name of the stream.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Log Streams using `log_group_name` and `name` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_cloudwatch_log_stream.example
  id = "example-group:example-stream"
}
```

Using `terraform import`, import Log Streams using `log_group_name` and `name` separated by a colon (`:`). For example:

```console
% terraform import aws_cloudwatch_log_stream.example example-group:example-stream
```
