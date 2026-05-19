---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_hosts"
description: |-
    Provides details about multiple EC2 Dedicated Hosts
---

# Data Source: aws_ec2_hosts

Provides a list of EC2 Dedicated Host IDs matching the provided filters. More information about Dedicated Hosts can be found in the [EC2 User Guide](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/dedicated-hosts-overview.html).

## Example Usage

### Filter by instance type

```terraform
data "aws_ec2_hosts" "example" {
  filter {
    name   = "instance-type"
    values = ["c5.large"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}
```

### Filter by Outpost ARN

The `outpost_arn` argument applies a client-side filter because the `DescribeHosts` API does not support `outpost-arn` as a server-side filter.

```terraform
data "aws_ec2_hosts" "outpost" {
  outpost_arn = data.aws_outposts_outpost.example.arn

  filter {
    name   = "state"
    values = ["available"]
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeHosts.html) for supported filters. Detailed below.
* `outpost_arn` - (Optional) ARN of the AWS Outpost. Filters results client-side to only include hosts allocated on this Outpost.
* `tags` - (Optional) Key-value map of resource tags, each pair of which must exactly match a pair on the desired Dedicated Hosts.

### filter Argument Reference

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of EC2 Dedicated Host identifiers.
