# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "address" {
  value = "Instances: ${element(aws_instance.web[*].id, 0)}"
}

