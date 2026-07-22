resource "aws_prometheus_workspace" "test" {
{{- template "region" }}
}

resource "aws_prometheus_anomaly_detector" "test" {
{{- template "region" }}
  alias = var.rName
  workspace_id = aws_prometheus_workspace.test.id

  configuration {
	random_cut_forest {
	  query = "avg(up)"
	}
  }
{{- template "tags" . }}
}