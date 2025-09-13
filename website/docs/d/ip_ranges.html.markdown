---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_ip_ranges"
description: |-
  Get information on AWS IP ranges.
---

# Data Source: aws_ip_ranges

Use this data source to get the IP ranges of various AWS products and services. For more information about the contents of this data source and required JSON syntax if referencing a custom URL, see the [AWS IP Address Ranges documentation][1].

## Example Usage

```terraform
data "aws_ip_ranges" "european_ec2" {
  regions  = ["eu-west-1", "eu-central-1"]
  services = ["ec2"]
}

resource "aws_security_group" "from_europe" {
  name = "from_europe"

  ingress {
    from_port        = "443"
    to_port          = "443"
    protocol         = "tcp"
    cidr_blocks      = data.aws_ip_ranges.european_ec2.cidr_blocks
    ipv6_cidr_blocks = data.aws_ip_ranges.european_ec2.ipv6_cidr_blocks
  }

  tags = {
    CreateDate = data.aws_ip_ranges.european_ec2.create_date
    SyncToken  = data.aws_ip_ranges.european_ec2.sync_token
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `regions` - (Optional) Filter IP ranges by regions (or include all regions, if
omitted). Valid items are `global` (for `cloudfront`) as well as all AWS regions
(e.g., `eu-central-1`)
* `services` - (Required) Filter IP ranges by services. Valid items are `amazon`
(for amazon.com), `amazon_connect`, `api_gateway`, `cloud9`, `cloudfront`,
`codebuild`, `dynamodb`, `ec2`, `ec2_instance_connect`, `globalaccelerator`,
`route53`, `route53_healthchecks`, `s3` and `workspaces_gateways`. See the
[`service` attribute][2] documentation for other possible values.

~> **NOTE:** If the specified combination of regions and services does not yield any
CIDR blocks, Terraform will fail.

* `url` - (Optional) Custom URL for source JSON file. Syntax must match [AWS IP Address Ranges documentation][1]. Defaults to `https://ip-ranges.amazonaws.com/ip-ranges.json`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cidr_blocks` - Lexically ordered list of CIDR blocks.
* `ipv6_cidr_blocks` - Lexically ordered list of IPv6 CIDR blocks.
* `create_date` - Publication time of the IP ranges (e.g., `2016-08-03-23-46-05`).
* `sync_token` - Publication time of the IP ranges, in Unix epoch time format
  (e.g., `1470267965`).

[1]: https://docs.aws.amazon.com/general/latest/gr/aws-ip-ranges.html
[2]: https://docs.aws.amazon.com/general/latest/gr/aws-ip-ranges.html#aws-ip-syntax
