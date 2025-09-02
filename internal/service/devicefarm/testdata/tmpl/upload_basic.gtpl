resource "aws_devicefarm_upload" "test" {
{{- template "region" }}
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
  type        = "APPIUM_JAVA_TESTNG_TEST_SPEC"
}

resource "aws_devicefarm_project" "test" {
{{- template "region" }}
  name = var.rName
}
