resource "aws_resiliencehubv2_policy" "test" {
{{- template "region" }}
  name = var.rName

  availability_slo {
    target = 99.9
  }

{{- template "tags" . }}
}
