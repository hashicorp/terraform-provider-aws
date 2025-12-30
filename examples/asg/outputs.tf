# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

output "security_group" {
  value = aws_security_group.default.id
}

output "launch_configuration" {
  value = aws_launch_template.web-lt.id
}

output "asg_name" {
  value = aws_autoscaling_group.web-asg.id
}

output "elb_name" {
  value = aws_elb.web-elb.dns_name
}
