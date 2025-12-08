---
subcategory: "CloudWatch NetworkFlow Monitor"
layout: "aws"
page_title: "AWS: aws_networkflowmonitor_monitor"
description: |-
  Manages a Network Flow Monitor Monitor.
---

# Resource: aws_networkflowmonitor_monitor

Manages a Network Flow Monitor Monitor.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example"
  }
}

resource "aws_networkflowmonitor_monitor" "example" {
  monitor_name = "example-monitor"
  scope_arn    = aws_networkflowmonitor_scope.example.scope_arn

  local_resource {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.example.arn
  }

  remote_resource {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.example.arn
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `monitor_name` - (Required) The name of the monitor. Cannot be changed after creation.
* `scope_arn` - (Required) The Amazon Resource Name (ARN) of the scope for the monitor. Cannot be changed after creation.

The following arguments are optional:

* `local_resource` - (Optional) The local resources to monitor. A local resource in a workload is the location of the hosts where the Network Flow Monitor agent is installed.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `remote_resource` - (Optional) The remote resources to monitor. A remote resource is the other endpoint specified for the network flow of a workload, with a local resource.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### local_resource and remote_resource

The `local_resource` and `remote_resource` blocks support the following:

* `type` - (Required) The type of the resource. Valid values are `AWS::EC2::VPC`, `AWS::EC2::Subnet`, `AWS::EC2::AvailabilityZone`, `AWS::EC2::Region`, and `AWS::EKS::Cluster`.
* `identifier` - (Required) The identifier of the resource. For VPC resources, this is the VPC ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `monitor_arn` - The Amazon Resource Name (ARN) of the monitor.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Flow Monitor Monitor using the monitor name. For example:

```terraform
import {
  to = aws_networkflowmonitor_monitor.example
  id = "example-monitor"
}
```

Using `terraform import`, import Network Flow Monitor Monitor using the monitor name. For example:

```console
% terraform import aws_networkflowmonitor_monitor.example example-monitor
```
