---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_connector"
description: |-
  Manages an AWS Config connector to a third-party cloud service provider.
---

# Resource: aws_config_connector

Manages an AWS Config connector that specifies the connection between a third-party cloud service provider and AWS Config.

A connector is a prerequisite for recording third-party (multi-cloud) resource configurations. For example, it provides the `arn` referenced by the `azure` provider block of [`aws_securityhub_connector_v2`](securityhub_connector_v2.html) to enable Microsoft Azure coverage in Security Hub.

~> **NOTE:** Creating this resource also creates the `AWSServiceRoleForConfigThirdParty` service-linked role in your account if it does not already exist.

~> **NOTE:** Connectors cannot be updated in place. Changing the `azure` configuration forces a new resource to be created.

Before creating a connector, you must configure the corresponding application and federated credentials in your Azure environment. See the [Security Hub Azure integration prerequisites](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-v2-azure-prereqs.html) for details.

## Example Usage

```terraform
resource "aws_config_connector" "example" {
  azure {
    client_identifier = "00000000-0000-0000-0000-000000000000"
    tenant_identifier = "11111111-1111-1111-1111-111111111111"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `azure` - (Required) Configuration for connecting to a Microsoft Azure environment. Changing this forces a new resource to be created. [See below](#azure).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### azure

The following arguments are supported in the `azure` block. Changing any of them forces a new resource to be created.

* `client_identifier` - (Required) Azure application (client) identifier that AWS uses to authenticate to the Azure environment.
* `tenant_identifier` - (Required) Azure (Microsoft Entra ID) tenant identifier.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector.
* `name` - Name of the connector.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_config_connector.example
  identity = {
    arn = "arn:aws:config:us-east-1:123456789012:connector/00000000-0000-0000-0000-000000000000"
  }
}

resource "aws_config_connector" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the connector.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Config connectors using the `arn`. For example:

```terraform
import {
  to = aws_config_connector.example
  id = "arn:aws:config:us-east-1:123456789012:connector/00000000-0000-0000-0000-000000000000"
}
```

Using `terraform import`, import Config connectors using the `arn`. For example:

```console
% terraform import aws_config_connector.example arn:aws:config:us-east-1:123456789012:connector/00000000-0000-0000-0000-000000000000
```
