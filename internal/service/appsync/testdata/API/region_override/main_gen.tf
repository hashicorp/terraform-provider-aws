# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appsync_api" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
