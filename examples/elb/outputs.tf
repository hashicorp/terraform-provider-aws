# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "address" {
  value = aws_elb.web.dns_name
}
