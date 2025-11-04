# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "TERRAFORMCLOUD"
  auth_type      = "apiKey"

  credential {
    api_key_credential {
      api_key = "apiexamplekeytest"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }

  tags = var.resource_tags
}

resource "aws_appfabric_app_bundle" "test" {}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
