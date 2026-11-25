resource "aws_cloudwatch_log_account_policy" "test" {
{{- template "region" }}
  policy_name = var.rName
  policy_type = "DATA_PROTECTION_POLICY"

  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {}
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}

data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name              = var.rName
  retention_in_days = 1
}
