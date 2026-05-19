resource "aws_xray_indexing_rule" "test" {
{{- template "region" }}
  name = var.rName

  rule {
    probabilistic {
      desired_sampling_percentage = 0.66
    }
  }
}
