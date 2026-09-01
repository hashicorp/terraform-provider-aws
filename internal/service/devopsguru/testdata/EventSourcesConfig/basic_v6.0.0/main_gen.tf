# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsguru_event_sources_config" "test" {
  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.0.0"
    }
  }
}

provider "aws" {}
