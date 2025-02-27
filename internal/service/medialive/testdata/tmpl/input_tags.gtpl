resource "aws_medialive_input" "test" {
  name                  = var.rName
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"
{{- template "tags" . }}
}

# testAccInputBaseConfig

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }
}
