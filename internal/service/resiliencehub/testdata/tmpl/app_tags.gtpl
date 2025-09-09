resource "aws_resiliencehub_app" "test" {
  name                    = var.rName
  app_assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "test-resource"
      type = "AWS::Lambda::Function"
      logical_resource_id {
        identifier = "TestResource"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "test-component"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["test-resource"]
    }
  }

{{- template "tags" . }}
}
