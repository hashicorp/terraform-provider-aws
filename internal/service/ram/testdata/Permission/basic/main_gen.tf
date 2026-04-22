# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ram_permission" "test" {
  name            = var.rName
  policy_template = <<EOF
{
    "Effect": "Allow",
    "Action": [
	"backup:ListProtectedResourcesByBackupVault",
	"backup:ListRecoveryPointsByBackupVault",
	"backup:DescribeRecoveryPoint",
	"backup:DescribeBackupVault"
    ]
}
EOF
  resource_type   = "backup:BackupVault"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
