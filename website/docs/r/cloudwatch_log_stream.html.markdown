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

* `name` - (Required) The name of the log stream. Must not be longer than 512 characters and must not contain `:`
* `log_group_name` - (Required) The name of the log group under which the log stream is to be created.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) specifying the log stream.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudwatch Log Stream using the stream's `log_group_name` and `name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_stream.foo
  id = "Yada:SampleLogStream1234"
}
```

Using `terraform import`, import Cloudwatch Log Stream using the stream's `log_group_name` and `name`. For example:

```console
% terraform import aws_cloudwatch_log_stream.foo Yada:SampleLogStream1234
```
