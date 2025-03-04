resource "aws_prometheus_workspace" "demo" {
}

resource "aws_prometheus_rule_group_namespace" "demo" {
  name         = "rules"
  workspace_id = aws_prometheus_workspace.demo.id
  {{- template "tags" . }}
  data         = <<EOF
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
EOF
}
