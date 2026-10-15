resource "aws_cloudwatch_event_connection" "test" {
{{- template "region" }}
  name               = var.rName
  authorization_type = "BASIC"

  auth_parameters {
    basic {
      username = "${var.rName}-user"
      password = "${var.rName}-pass"
    }
  }
}
