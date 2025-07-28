resource "aws_iam_virtual_mfa_device" "test" {
  virtual_mfa_device_name = var.rName
{{- template "tags" . }}
}
