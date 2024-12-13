---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_network_insights_path"
description: |-
  Provides a Network Insights Path resource.
---

# Resource: aws_ec2_network_insights_path

Provides a Network Insights Path resource. Part of the "Reachability Analyzer" service in the AWS VPC console.

## Example Usage

```terraform
resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.source.id
  destination = aws_network_interface.destination.id
  protocol    = "tcp"
}
```

## Argument Reference

The following arguments are required:

* `source` - (Required) ID or ARN of the resource which is the source of the path. Can be an Instance, Internet Gateway, Network Interface, Transit Gateway, VPC Endpoint, VPC Peering Connection or VPN Gateway. If the resource is in another account, you must specify an ARN.
* `destination` - (Optional) ID or ARN of the resource which is the destination of the path. Can be an Instance, Internet Gateway, Network Interface, Transit Gateway, VPC Endpoint, VPC Peering Connection or VPN Gateway. If the resource is in another account, you must specify an ARN.
* `protocol` - (Required) Protocol to use for analysis. Valid options are `tcp` or `udp`.

The following arguments are optional:

* `source_ip` - (Optional) IP address of the source resource.
* `destination_ip` - (Optional) IP address of the destination resource.
* `destination_port` - (Optional) Destination port to analyze access to.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Network Insights Path.
* `destination_arn` - ARN of the destination.
* `id` - ID of the Network Insights Path.
* `source_arn` - ARN of the source.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Insights Paths using the `id`. For example:

```terraform
import {
  to = aws_ec2_network_insights_path.test
  id = "nip-00edfba169923aefd"
}
```

Using `terraform import`, import Network Insights Paths using the `id`. For example:

```console
% terraform import aws_ec2_network_insights_path.test nip-00edfba169923aefd
```
