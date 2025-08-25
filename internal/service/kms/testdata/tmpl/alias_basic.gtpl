resource "aws_kms_alias" "test" {
{{- template "region" }}
  name          = "alias/${var.rName}"
  target_key_id = aws_kms_key.test.id
}

resource "aws_kms_key" "test" {
{{- template "region" }}
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
