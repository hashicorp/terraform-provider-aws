---
subcategory: "SSM Quick Setup"
layout: "aws"
page_title: "AWS: aws_ssmquicksetup_configuration_manager"
description: |-
  Terraform resource for managing an AWS SSM Quick Setup Configuration Manager.
---
# Resource: aws_ssmquicksetup_configuration_manager

Terraform resource for managing an AWS SSM Quick Setup Configuration Manager.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssmquicksetup_configuration_manager" "example" {
  name = "example"

  configuration_definition {
    type                                    = ""
    local_deployment_administrator_role_arn = aws_iam_role.example.arn
    local_deployment_execution_role_name    = aws_iam_role.example.name

    parameters = {
      "key" = "value"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration_definition` - (Required) Definition of the Quick Setup configuration that the configuration manager deploys. See [`configuration_definition`](#configuration_definition-argument-reference) below.
* `name` - (Required) Configuration manager name.

The following arguments are optional:

* `description` - (Optional) Description of the configuration manager.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `configuration_definition` Argument Reference

* `local_deployment_administrator_role_arn` - (Optional) ARN of the IAM role used to administrate local configuration deployments.
* `local_deployment_execution_role_name` - (Optional) Name of the IAM role used to deploy local configurations.
* `parameters` - (Required) Parameters for the configuration definition type. Parameters for configuration definitions vary based the configuration type. See the [AWS API documentation](https://docs.aws.amazon.com/quick-setup/latest/APIReference/API_ConfigurationDefinitionInput.html) for a complete list of parameters for each configuration type.
* `type` - (Required) Type of the Quick Setup configuration.
* `type_version` - (Optional) Version of the Quick Setup type to use.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `manager_arn` - ARN of the Configuration Manager.
* `status_summaries` - A summary of the state of the configuration manager. This includes deployment statuses, association statuses, drift statuses, health checks, and more. See [`status_summaries`](#status_summaries-attribute-reference) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `status_summaries` Attribute Reference

* `status` - Current status.
* `status_message` - When applicable, returns an informational message relevant to the current status and status type of the status summary object.
* `status_type` - Type of a status summary.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Quick Setup Configuration Manager using the `maanger_arn`. For example:

```terraform
import {
  to = aws_ssmquicksetup_configuration_manager.example
  id = "arn:aws:ssm-quicksetup:us-east-1:012345678901:configuration-manager/abcd-1234"
}
```

Using `terraform import`, import SSM Quick Setup Configuration Manager using the `maanger_arn`. For example:

```console
% terraform import aws_ssmquicksetup_configuration_manager.example arn:aws:ssm-quicksetup:us-east-1:012345678901:configuration-manager/abcd-1234
```
