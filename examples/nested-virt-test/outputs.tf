# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "instance_id" {
  value = aws_instance.nested_virt.id
}

output "cpu_options" {
  value = aws_instance.nested_virt.cpu_options
}
