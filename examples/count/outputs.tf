# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "address" {
  value = "Instances: ${element(aws_instance.web[*].id, 0)}"
}

