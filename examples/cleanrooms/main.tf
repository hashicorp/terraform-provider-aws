# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_cleanrooms_collaboration" "test_collab" {
  name                     = "terraform-example-collaboration"
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = "Creator "
  description              = "I made this collaboration with terraform!"
  query_log_status         = "DISABLED"

  data_encryption_metadata {
    allow_clear_text                            = true
    allow_duplicates                            = true
    allow_joins_on_columns_with_different_names = true
    preserve_nulls                              = false
  }

  member {
    account_id       = 123456789012
    display_name     = "Other member"
    member_abilities = []
  }

  tags = {
    Project = "Terraform"
  }

}
