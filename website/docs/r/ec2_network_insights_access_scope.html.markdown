---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_network_insights_access_scope"
description: |-
  Provides a Network Insights Access Scope resource.
---

# Resource: aws_ec2_network_insights_access_scope

Provides a Network Insights Access Scope resource.
Part of the "Network Access Analyzer" service in the AWS VPC console.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_network_insights_access_scope" "example" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
    destination {
      resource_statement {
        resource_types = ["AWS::EC2::InternetGateway"]
      }
    }
  }
}
```

### With Exclude Paths

```terraform
resource "aws_ec2_network_insights_access_scope" "example" {
  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }

  exclude_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::InternetGateway"]
      }
    }
    through_resources {
      resource_statement {
        resource_types = ["AWS::EC2::NatGateway"]
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `match_paths` - (Required) Set of access scope path statements to match.
  At least one must be specified.
  See [`match_paths`](#match_paths) below for details.

The following arguments are optional:

* `exclude_paths` - (Optional) Set of access scope path statements to exclude.
  See [`exclude_paths`](#exclude_paths) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints).
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource.
  If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### match_paths

* `source` - (Optional) Path statement for the source.
  See [`source` and `destination`](#source-and-destination) below for details.
* `destination` - (Optional) Path statement for the destination.
  See [`source` and `destination`](#source-and-destination) below for details.

### exclude_paths

* `source` - (Optional) Path statement for the source.
  See [`source` and `destination`](#source-and-destination) below for details.
* `destination` - (Optional) Path statement for the destination.
  See [`source` and `destination`](#source-and-destination) below for details.
* `through_resources` - (Optional) Path statement for through resources.
  See [`through_resources`](#through_resources) below for details.

### source and destination

* `packet_header_statement` - (Optional) Packet header statement.
  See [`packet_header_statement`](#packet_header_statement) below for details.
* `resource_statement` - (Optional) Resource statement.
  Exactly one of `resources` or `resource_types` must be specified.
  See [`resource_statement`](#resource_statement) below for details.

### through_resources

* `resource_statement` - (Optional) Resource statement.
  Exactly one of `resources` or `resource_types` must be specified.
  See [`resource_statement`](#resource_statement) below for details.

### packet_header_statement

* `source_addresses` - (Optional) Set of source addresses.
* `destination_addresses` - (Optional) Set of destination addresses.
* `source_ports` - (Optional) Set of source ports.
* `destination_ports` - (Optional) Set of destination ports.
* `source_prefix_lists` - (Optional) Set of source prefix lists.
* `destination_prefix_lists` - (Optional) Set of destination prefix lists.
* `protocols` - (Optional) Set of protocols.
  Valid values are `tcp` and `udp`.

### resource_statement

* `resources` - (Optional) List of resource ARNs.
  Cannot be specified together with `resource_types`.
* `resource_types` - (Optional) List of resource types.
  Cannot be specified together with `resources`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Network Insights Access Scope.
* `id` - ID of the Network Insights Access Scope.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_network_insights_access_scope.example
  identity = {
    id = "nis-0a1b2c3d4e5f6g7h8"
  }
}

resource "aws_ec2_network_insights_access_scope" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the Network Insights Access Scope.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Insights Access Scopes using the `id`. For example:

```terraform
import {
  to = aws_ec2_network_insights_access_scope.example
  id = "nis-0a1b2c3d4e5f6g7h8"
}
```

Using `terraform import`, import Network Insights Access Scopes using the `id`. For example:

```console
% terraform import aws_ec2_network_insights_access_scope.example nis-0a1b2c3d4e5f6g7h8
```
