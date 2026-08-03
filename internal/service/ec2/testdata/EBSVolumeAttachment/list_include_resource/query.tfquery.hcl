# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_volume_attachment" "test" {
  provider = aws

  config {
    instance_id = aws_instance.test.id
  }

  include_resource = true
}
