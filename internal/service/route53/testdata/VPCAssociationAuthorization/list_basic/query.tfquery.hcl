# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_route53_vpc_association_authorization" "test" {
  provider = aws

  config {
    zone_id = aws_route53_zone.test.id
  }
}
