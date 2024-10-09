# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "room" {
  value = aws_ivschat_room.example.arn
}
