resource "aws_kms_replica_external_key" "test" {
  description             = var.rName
  enabled                 = true
  primary_key_arn         = aws_kms_external_key.test.arn
  deletion_window_in_days = 7

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

{{- template "tags" . }}
}

# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = "${var.rName}-source"
  multi_region = true
  enabled      = true

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  deletion_window_in_days = 7
}
