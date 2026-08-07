resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_networkfirewall_container_association" "test" {
{{- template "region" }}
  container_association_name = var.rName
  type                       = "ECS"

  container_monitoring_configurations {
    cluster_arn = aws_ecs_cluster.test.arn
  }
{{- template "tags" . }}
}
