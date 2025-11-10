resource "aws_prometheus_rule_group_namespace" "test" {
{{- template "region" }}
  name         = var.rName
  workspace_id = aws_prometheus_workspace.test.id
  data         = <<EOF
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
EOF
{{- template "tags" . }}
}

resource "aws_prometheus_workspace" "test" {
{{- template "region" }}
}
