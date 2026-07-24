resource "aws_opensearchserverless_lifecycle_policy" "test" {
{{- template "region" }}
  name = var.rName
  type = "retention"
  policy = jsonencode({
    Rules : [
      {
        ResourceType : "index",
        Resource : ["index/${var.rName}/*"],
        MinIndexRetention : "81d"
      },
      {
        ResourceType : "index",
        Resource : ["index/sales/${var.rName}*"],
        NoMinIndexRetention : true
      }
    ]
  })

{{- template "tags" . }}
}
