resource "aws_devicefarm_device_pool" "test" {
{{- template "region" }}
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
  rule {
    attribute = "OS_VERSION"
    operator  = "EQUALS"
    value     = "\"AVAILABLE\""
  }
{{- template "tags" . }}
}

# testAccProjectConfig_basic

resource "aws_devicefarm_project" "test" {
{{- template "region" }}
  name = var.rName
}
