data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
{{- template "region" }}
  name = var.rName
  type = "data"
  policy = jsonencode([
    {
      Rules : [
        {
          ResourceType : "index",
          Resource : [
            "index/books/*"
          ],
          Permission : [
            "aoss:CreateIndex",
            "aoss:ReadDocument",
            "aoss:UpdateIndex",
            "aoss:DeleteIndex",
            "aoss:WriteDocument"
          ]
        }
      ],
      Principal : [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin"
      ]
    }
  ])

{{- template "tags" . }}
}
