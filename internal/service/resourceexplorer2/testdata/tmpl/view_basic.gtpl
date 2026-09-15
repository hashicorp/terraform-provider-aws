resource "aws_resourceexplorer2_view" "test" {
{{- template "region" }}
  name = var.rName

  depends_on = [aws_resourceexplorer2_index.test]
{{- template "tags" . }}
}

resource "aws_resourceexplorer2_index" "test" {
{{- template "region" }}
  type = "LOCAL"
}
