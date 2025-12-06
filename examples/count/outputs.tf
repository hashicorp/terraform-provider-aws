# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

output "address" {
  value = "Instances: ${element(aws_instance.web[*].id, 0)}"
}

