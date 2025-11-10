resource "aws_appsync_api" "test" {
  {{- template "region" }}
  name = var.rName

  event_config {
    auth_provider {
      auth_type = "API_KEY"
    }

    connection_auth_mode {
      auth_type = "API_KEY"
    }

    default_publish_auth_mode {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_mode {
      auth_type = "API_KEY"
    }
  }
  {{- template "tags" . }}
}

resource "aws_appsync_channel_namespace" "test" {
  {{- template "region" }}
  api_id = aws_appsync_api.test.api_id
  name   = var.rName

  {{- template "tags" . }}
}
