resource "aws_networksecuritymanager_scope" "test" {
{{- template "region" }}
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
      include_all   = true
    }
  }
{{- template "tags" . }}
}
