---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_service_action"
description: |-
  Manages a Service Catalog Service Action
---

# Resource: aws_servicecatalog_service_action

Manages a Service Catalog self-service action.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_service_action" "example" {
  description = "Motor generator unit"
  name        = "MGU"

  definition {
    name = "AWS-RestartEC2Instance"
  }
}
```

## Argument Reference

The following arguments are required:

* `definition` - (Required) Self-service action definition configuration block. Detailed below.
* `name` - (Required) Self-service action name.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values are `en` (English), `jp` (Japanese), and `zh` (Chinese). Default is `en`.
* `description` - (Optional) Self-service action description.

### `definition`

The `definition` configuration block supports the following attributes:

* `assume_role` - (Optional) ARN of the role that performs the self-service actions on your behalf. For example, `arn:aws:iam::12345678910:role/ActionRole`. To reuse the provisioned product launch role, set to `LAUNCH_ROLE`.
* `name` - (Required) Name of the SSM document. For example, `AWS-RestartEC2Instance`. If you are using a shared SSM document, you must provide the ARN instead of the name.
* `parameters` - (Optional) List of parameters in JSON format. For example: `[{\"Name\":\"InstanceId\",\"Type\":\"TARGET\"}]` or `[{\"Name\":\"InstanceId\",\"Type\":\"TEXT_VALUE\"}]`.
* `type` - (Optional) Service action definition type. Valid value is `SSM_AUTOMATION`. Default is `SSM_AUTOMATION`.
* `version` - (Required) SSM document version. For example, `1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the service action.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `10m`)
- `update` - (Default `3m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_service_action` using the service action ID. For example:

```terraform
import {
  to = aws_servicecatalog_service_action.example
  id = "act-f1w12eperfslh"
}
```

Using `terraform import`, import `aws_servicecatalog_service_action` using the service action ID. For example:

```console
% terraform import aws_servicecatalog_service_action.example act-f1w12eperfslh
```
