---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_channel_namespace"
description: |-
  Manages an AWS AppSync Channel Namespace.
---

# Resource: aws_appsync_channel_namespace

Manages an [AWS AppSync Channel Namespace](https://docs.aws.amazon.com/appsync/latest/eventapi/event-api-concepts.html#namespace).

## Example Usage

### Basic Usage

```terraform
resource "aws_appsync_channel_namespace" "example" {
  name   = "example-channel-namespace"
  api_id = aws_appsync_api.example.api_id
}
```

## Argument Reference

The following arguments are required:

* `api_id` - (Required) Event API ID.
* `name` - (Required) Name of the channel namespace.

The following arguments are optional:

* `code_handlers` - (Optional) Event handler functions that run custom business logic to process published events and subscribe requests.
* `handler_configs` - (Optional) Configuration for the `on_publish` and `on_subscribe` handlers. See [`handler_configs`](#handler_configs-block) below.
* `publish_auth_mode` - (Optional) Authorization modes to use for publishing messages on the channel namespace. This configuration overrides the default API authorization configuration. See [`publish_auth_mode`](#publish_auth_mode-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subscribe_auth_mode` - (Optional) Authorization modes to use for subscribing to messages on the channel namespace. This configuration overrides the default API authorization configuration. See [`subscribe_auth_mode`](#subscribe_auth_mode-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `handler_configs` Block

* `on_publish` - (Optional) Handler configuration for published events. See [`on_publish`](#on_publish-block) below.
* `on_subscribe` - (Optional) Handler configuration for subscribe requests. See [`on_subscribe`](#on_subscribe-block) below.

### `on_publish` Block

* `behavior` - (Required) Behavior for the handler. Valid values: `CODE`, `DIRECT`.
* `integration` - (Required) Integration data source configuration for the handler. See [`integration`](#integration-block) below.

### `on_subscribe` Block

* `behavior` - (Required) Behavior for the handler. Valid values: `CODE`, `DIRECT`.
* `integration` - (Required) Integration data source configuration for the handler. See [`integration`](#integration-block) below.

### `integration` Block

* `data_source_name` - (Required) Unique name of the data source that has been configured on the API.
* `lambda_config` - (Optional) Configuration for a Lambda data source. See [`lambda_config`](#lambda_config-block) below.

### `lambda_config` Block

* `invoke_type` - (Optional) Invocation type for a Lambda data source. Valid values: `REQUEST_RESPONSE`, `EVENT`.

### `publish_auth_mode` Block

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

### `subscribe_auth_mode` Block

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `channel_namespace_arn` - ARN of the channel namespace.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppSync Channel Namespace using the `api_id` and `name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_appsync_channel_namespace.example
  id = "example-api-id,example-channel-namespace"
}
```

Using `terraform import`, import AppSync Channel Namespace using the `api_id` and `name` separated by a comma (`,`). For example:

```console
% terraform import aws_appsync_channel_namespace.example example-api-id,example-channel-namespace
```
