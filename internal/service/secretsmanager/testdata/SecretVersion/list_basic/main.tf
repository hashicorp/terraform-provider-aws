# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_secretsmanager_secret_version" "test" {
  count = 2

  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string-${count.index}"
}

resource "aws_secretsmanager_secret" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
