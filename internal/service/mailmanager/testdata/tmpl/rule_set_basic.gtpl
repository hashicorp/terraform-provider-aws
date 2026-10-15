resource "aws_mailmanager_rule_set" "test" {
{{- template "region" }}
  name = var.rName

  rule {
    name = "example"

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }
  }
{{- template "tags" . }}
}
