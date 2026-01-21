resource "aws_opensearchserverless_security_policy" "test" {
{{- template "region" }}
  name = var.rName
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/${var.rName}"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" }}

  depends_on = [aws_opensearchserverless_security_policy.test]
}
