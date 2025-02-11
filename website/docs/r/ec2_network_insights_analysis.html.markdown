---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_network_insights_analysis"
description: |-
  Provides a Network Insights Analysis resource.
---

# Resource: aws_ec2_network_insights_analysis

Provides a Network Insights Analysis resource. Part of the "Reachability Analyzer" service in the AWS VPC console.

## Example Usage

```terraform
resource "aws_ec2_network_insights_path" "path" {
  source      = aws_network_interface.source.id
  destination = aws_network_interface.destination.id
  protocol    = "tcp"
}

resource "aws_ec2_network_insights_analysis" "analysis" {
  network_insights_path_id = aws_ec2_network_insights_path.path.id
}
```

## Argument Reference

The following arguments are required:

* `network_insights_path_id` - (Required) ID of the Network Insights Path to run an analysis on.

The following arguments are optional:

* `filter_in_arns` - (Optional) A list of ARNs for resources the path must traverse.
* `wait_for_completion` - (Optional) If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alternate_path_hints` - Potential intermediate components of a feasible path. Described below.
* `arn` - ARN of the Network Insights Analysis.
* `explanations` - Explanation codes for an unreachable path. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Explanation.html) for details.
* `forward_path_components` - The components in the path from source to destination. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
* `id` - ID of the Network Insights Analysis.
* `path_found` - Set to `true` if the destination was reachable.
* `return_path_components` - The components in the path from destination to source. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
* `start_date` - The date/time the analysis was started.
* `status` - The status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `path_found`.
* `status_message` - A message to provide more context when the `status` is `failed`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `warning_message` - The warning message.

The `alternate_path_hints` object supports the following:

* `component_arn` - The Amazon Resource Name (ARN) of the component.
* `component_id` - The ID of the component.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Insights Analyzes using the `id`. For example:

```terraform
import {
  to = aws_ec2_network_insights_analysis.test
  id = "nia-0462085c957f11a55"
}
```

Using `terraform import`, import Network Insights Analyzes using the `id`. For example:

```console
% terraform import aws_ec2_network_insights_analysis.test nia-0462085c957f11a55
```
