resource "aws_networksecuritymanager_scope" "test" {
{{- template "region" }}
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            tags = {
              Name = var.rName
            }
          }
        }
      }
    }
  }
{{- template "tags" . }}
}
