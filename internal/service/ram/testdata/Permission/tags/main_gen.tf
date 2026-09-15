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

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
