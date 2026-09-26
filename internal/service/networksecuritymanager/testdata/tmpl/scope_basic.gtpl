resource "aws_networksecuritymanager_scope" "test" {
{{- template "region" }}
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = var.rName
            }
          }
        }
      }
    }
  }
{{- template "tags" . }}
}
