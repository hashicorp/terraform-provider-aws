---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_prompt"
description: |-
  Provides details about a specific Amazon Connect Prompt
---

# Resource: aws_connect_prompt

Provides an Amazon Connect Prompt resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

### Basic

```terraform
resource "aws_connect_prompt" "example" {
  instance_id = aws_connect_instance.example.id
  name         = "example"
  description  = "example"
  s3_uri       = "s3://${aws_s3_object.example.bucket}/sample.wav"
}
```

## Argument Reference
* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `name` - (Required) Specifies the name of the Prompt.
* `description` - (Optional) Specifies the description of the Prompt.
* `s3_uri` - (Required) The Amazon S3 URI (`https://` or `s3://`) for the prompt.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Prompt.
* `prompt_id` - The identifier for the Prompt.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the Quick Connect separated by a colon (`:`).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Prompt using the `instance_id` and `prompt_id` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_connect_prompt.example
  id = "7695b813-83ae-4cbc-895f-381c9503a730:0840329b-56d7-4902-988f-1505cfc730f3"
}
```

Using `terraform import`, import Amazon Connect Prompt using the `instance_id` and `prompt_id` separated by a colon (`:`). For example:

```console
% terraform import aws_connect_prompt.example 7695b813-83ae-4cbc-895f-381c9503a730:0840329b-56d7-4902-988f-1505cfc730f3
```
