resource "aws_prometheus_workspace" "test" {
}

resource "aws_prometheus_rule_group_namespace" "test" {
  name         = var.rName
  workspace_id = aws_prometheus_workspace.test.id
{{- template "tags" . }}
  data = <<EOF
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
EOF
}
