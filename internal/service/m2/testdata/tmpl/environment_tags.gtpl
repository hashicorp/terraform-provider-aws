resource "aws_m2_environment" "test" {
  name          = var.rName
  engine_type   = "microfocus"
  instance_type = "M2.m5.large"
{{- template "tags" . }}
}
