# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_appsync_api" "test" {
  name = var.rName

  event_config {
    auth_providers {
      auth_type = "API_KEY"
    }

    connection_auth_modes {
      auth_type = "API_KEY"
    }

    default_publish_auth_modes {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_modes {
      auth_type = "API_KEY"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
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
