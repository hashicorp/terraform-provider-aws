---
layout: "aws"
page_title: "AWS: aws_eip"
sidebar_current: "docs-aws-datasource-eip"
description: |-
    Provides details about a specific Elastic IP
---

# Data Source: aws_eip

`aws_eip` provides details about a specific Elastic IP.

This resource can prove useful when a module accepts an allocation ID or
public IP as an input variable and needs to determine the other.

## Example Usage

The following example shows how one might accept a public IP as a variable
and use this data source to obtain the allocation ID.

```hcl
variable "instance_id" {}
variable "public_ip" {}

data "aws_eip" "proxy_ip" {
  public_ip = "${var.public_ip}"
}

resource "aws_eip_association" "proxy_eip" {
  instance_id   = "${var.instance_id}"
  allocation_id = "${data.aws_eip.proxy_ip.id}"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Elastic IPs in the current region. The given filters must match exactly one
Elastic IP whose data will be exported as attributes.

* `id` - (Optional) The allocation id of the specific EIP to retrieve.

* `public_ip` - (Optional) The public IP of the specific EIP to retrieve.

* `assocation_id` - (Optional) The Association ID of the specific EIP to retrieve.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAddresses.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes are also exported as result attributes. This
data source will complete the data by populating any fields that are not
included in the configuration with the data for the selected Elastic IP.

