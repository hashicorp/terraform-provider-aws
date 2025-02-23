resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.2.0.0/16"
  }
{{- template "tags" . }}
}
