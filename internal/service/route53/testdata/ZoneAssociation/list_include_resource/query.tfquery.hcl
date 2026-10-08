# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_route53_zone_association" "test" {
  provider = aws

  include_resource = true

  config {
    vpc_id = aws_vpc.bar.id
  }
}
