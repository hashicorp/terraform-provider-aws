---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_service_link_virtual_interfaces"
description: |-
    Provides details about multiple EC2 Service Link Virtual Interfaces
---

# Data Source: aws_ec2_service_link_virtual_interfaces

Provides a list of EC2 Service Link Virtual Interface IDs matching the provided filters. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/how-outposts-works.html#how-service-link).

## Example Usage

```terraform
data "aws_ec2_service_link_virtual_interfaces" "example" {
  filter {
    name   = "outpost-arn"
    values = [aws_outposts_outpost.example.arn]
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeServiceLinkVirtualInterfaces.html) for supported filters. Detailed below.
* `tags` - (Optional) Key-value map of resource tags, each pair of which must exactly match a pair on the desired Service Link Virtual Interfaces.

### filter Argument Reference

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of EC2 Service Link Virtual Interface identifiers.
