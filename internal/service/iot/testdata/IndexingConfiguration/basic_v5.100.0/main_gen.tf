# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iot_indexing_configuration" "test" {
  thing_group_indexing_configuration {
    thing_group_indexing_mode = "OFF"
  }

  thing_indexing_configuration {
    thing_indexing_mode = "OFF"
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.100.0"
    }
  }
}

provider "aws" {}
