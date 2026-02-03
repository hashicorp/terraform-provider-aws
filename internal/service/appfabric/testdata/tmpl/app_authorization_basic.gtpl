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
{{- template "tags" . }}
}

resource "aws_appfabric_app_bundle" "test" {}
