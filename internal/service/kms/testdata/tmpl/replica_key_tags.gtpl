resource "aws_kms_replica_key" "test" {
  description             = var.rName
  primary_key_arn         = aws_kms_key.test.arn
  deletion_window_in_days = 7

{{- template "tags" . }}
}

resource "aws_kms_key" "test" {
  provider = awsalternate

  description  = "${var.rName}-source"
  multi_region = true

  deletion_window_in_days = 7
}
