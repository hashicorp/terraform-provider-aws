resource "aws_devicefarm_network_profile" "test" {
{{- template "region" }}
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
{{- template "tags" . }}
}

# testAccProjectConfig_basic

resource "aws_devicefarm_project" "test" {
{{- template "region" }}
  name = var.rName
}
