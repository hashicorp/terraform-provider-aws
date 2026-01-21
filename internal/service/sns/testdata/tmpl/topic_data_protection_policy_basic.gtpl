data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_sns_topic_data_protection_policy" "test" {
{{- template "region" }}
  arn = aws_sns_topic.test.arn
  policy = jsonencode(
    {
      "Description" = "Default data protection policy"
      "Name"        = "__default_data_protection_policy"
      "Statement" = [
        {
          "DataDirection" = "Inbound"
          "DataIdentifier" = [
            "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
          ]
          "Operation" = {
            "Deny" = {}
          }
          "Principal" = [
            "*",
          ]
          "Sid" = var.rName
        },
      ]
      "Version" = "2021-06-01"
    }
  )
{{- template "tags" }}
}
