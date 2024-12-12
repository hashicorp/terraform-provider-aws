---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_network_insights_analysis"
description: |-
    Provides details about a specific Network Insights Analysis.
---

# Data Source: aws_ec2_network_insights_analysis

`aws_ec2_network_insights_analysis` provides details about a specific Network Insights Analysis.

## Example Usage

```terraform
data "aws_ec2_network_insights_analysis" "example" {
  network_insights_analysis_id = aws_ec2_network_insights_analysis.example.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Network Insights Analyzes. The given filters must match exactly one Network Insights Analysis
whose data will be exported as attributes.

* `network_insights_analysis_id` - (Optional) ID of the Network Insights Analysis to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the EC2 [`DescribeNetworkInsightsAnalyses`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNetworkInsightsAnalyses.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `alternate_path_hints` - Potential intermediate components of a feasible path.
* `arn` - ARN of the selected Network Insights Analysis.
* `explanations` - Explanation codes for an unreachable path.
* `filter_in_arns` - ARNs of the AWS resources that the path must traverse.
* `forward_path_components` - The components in the path from source to destination.
* `network_insights_path_id` - The ID of the path.
* `path_found` - Set to `true` if the destination was reachable.
* `return_path_components` - The components in the path from destination to source.
* `start_date` - Date/time the analysis was started.
* `status` - Status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `path_found`.
* `status_message` - Message to provide more context when the `status` is `failed`.
* `warning_message` - Warning message.
