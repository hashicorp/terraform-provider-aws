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
* `protocol` - (Required) Protocol to use for analysis. Valid options are `tcp` or `udp`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_ip` - (Optional) IP address of the source resource.
* `destination` - (Optional) ID or ARN of the resource which is the destination of the path. Can be an Instance, Internet Gateway, Network Interface, Transit Gateway, VPC Endpoint, VPC Peering Connection or VPN Gateway. If the resource is in another account, you must specify an ARN. Either the `destination` argument or the `destination_address` argument in the `filter_at_source` block must be specified.
* `destination_ip` - (Optional) IP address of the destination resource.
* `destination_port` - (Optional) Destination port to analyze access to.
* `filter_at_destination` - (Optional) Scopes the analysis to network paths that match specific filters at the destination. If you specify this parameter, you can't specify `destination_ip`. See below for details. Note that Terraform performs drift detection on this argument only when the value is provided.
* `filter_at_source` - (Optional) Scopes the analysis to network paths that match specific filters at the source. If you specify this parameter, you can't specify `source_ip` or `destination_port`. See below for details. Note that Terraform performs drift detection on this argument only when the value is provided.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `filter_at_destination` and `filter_at_source` configuration blocks support the following arguments:

* `destination_address` - (Optional) The destination IPv4 address.
* `destination_port_range` - (Optional) The destination port range. See below for details.
* `source_address` - (Optional) IP address of the source resource.
* `source_port_range` - (Optional) The source port range. See below for details.

The `destination_port_range` and `source_port_range` configuration blocks support the following arguments:

* `from_port` - (Optional) The first port in the range.
* `to_port` - (Optional) The last port in the range.

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
