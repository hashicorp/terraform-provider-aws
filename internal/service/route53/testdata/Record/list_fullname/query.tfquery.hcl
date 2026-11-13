# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_route53_record" "test" {
  provider = aws

  config {
    zone_id = aws_route53_zone.test.zone_id
  }
}
