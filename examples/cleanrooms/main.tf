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
  creator_member_abilities = ["CAN_QUERY"]
  creator_display_name     = "Creator"
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
    member_abilities = ["CAN_RECEIVE_RESULTS"]
  }

  tags = {
    Project = "Terraform"
  }
}

resource "aws_cleanrooms_membership" "test_membership" {
  collaboration_id = aws_cleanrooms_collaboration.test_collab.id
  query_log_status = "DISABLED"

  default_result_configuration {
    role_arn = "arn:aws:iam::123456789012:role/role-name"
    output_configuration {
      s3 {
        bucket        = "test-bucket"
        result_format = "PARQUET"
        key_prefix    = "test-prefix"
      }
    }
  }

  tags = {
    Project = "Terraform"
  }
}

resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "column1",
    "column2",
    "column3",
  ]

  table_reference {
    database_name = "example_database"
    table_name    = "example_table"
  }

  tags = {
    Project = "Terraform"
  }
}
