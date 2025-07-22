resource "aws_appsync_event_api" "test" {
  {{- template "region" }}
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
  {{- template "tags" . }}
}
