---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_agent"
description: |-
  Manages an AWS DataSync Agent in the provider region
---

# Resource: aws_datasync_agent

Manages an AWS DataSync Agent deployed on premises.

~> **NOTE:** One of `activation_key` or `ip_address` must be provided for resource creation (agent activation). Neither is required for resource import. If using `ip_address`, Terraform must be able to make an HTTP (port 80) GET request to the specified IP address from where it is running. The agent will turn off that HTTP server after activation.

## Example Usage

```terraform
resource "aws_datasync_agent" "example" {
  ip_address = "1.2.3.4"
  name       = "example"
}
```

## Example Usage with VPC Endpoints

```terraform
resource "aws_datasync_agent" "example" {
  ip_address            = "1.2.3.4"
  security_group_arns   = [aws_security_group.example.arn]
  subnet_arns           = [aws_subnet.example.arn]
  vpc_endpoint_id       = aws_vpc_endpoint.example.id
  private_link_endpoint = data.aws_network_interface.example.private_ip
  name                  = "example"
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "example" {
  service_name       = "com.amazonaws.${data.aws_region.current.region}.datasync"
  vpc_id             = aws_vpc.example.id
  security_group_ids = [aws_security_group.example.id]
  subnet_ids         = [aws_subnet.example.id]
  vpc_endpoint_type  = "Interface"
}

data "aws_network_interface" "example" {
  id = tolist(aws_vpc_endpoint.example.network_interface_ids)[0]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the DataSync Agent.
* `activation_key` - (Optional) DataSync Agent activation key during resource creation. Conflicts with `ip_address`. If an `ip_address` is provided instead, Terraform will retrieve the `activation_key` as part of the resource creation.
* `ip_address` - (Optional) DataSync Agent IP address to retrieve activation key during resource creation. Conflicts with `activation_key`. DataSync Agent must be accessible on port 80 from where Terraform is running.
* `private_link_endpoint` - (Optional) The IP address of the VPC endpoint the agent should connect to when retrieving an activation key during resource creation. Conflicts with `activation_key`.
* `security_group_arns` - (Optional) The ARNs of the security groups used to protect your data transfer task subnets.
* `subnet_arns` - (Optional) The Amazon Resource Names (ARNs) of the subnets in which DataSync will create elastic network interfaces for each data transfer task.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Agent. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_endpoint_id` - (Optional) The ID of the VPC (virtual private cloud) endpoint that the agent has access to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DataSync Agent.
* `arn` - Amazon Resource Name (ARN) of the DataSync Agent.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_agent` using the DataSync Agent Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_datasync_agent.example
  id = "arn:aws:datasync:us-east-1:123456789012:agent/agent-12345678901234567"
}
```

Using `terraform import`, import `aws_datasync_agent` using the DataSync Agent Amazon Resource Name (ARN). For example:

```console
% terraform import aws_datasync_agent.example arn:aws:datasync:us-east-1:123456789012:agent/agent-12345678901234567
```
