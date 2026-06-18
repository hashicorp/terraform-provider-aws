data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_resiliencehubv2_policy" "test" {
{{- template "region" }}
  name = "${var.rName}-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
{{- template "region" }}
  name    = "${var.rName}-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_service_function" "test" {
{{- template "region" }}
  name        = var.rName
  service_arn = aws_resiliencehubv2_service.test.arn
  criticality = "PRIMARY"
}
