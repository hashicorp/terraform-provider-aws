resource "aws_ssm_patch_group" "test" {
{{- template "region" }}
  baseline_id = aws_ssm_patch_baseline.test.id
  patch_group = var.rName
}

resource "aws_ssm_patch_baseline" "test" {
{{- template "region" }}
  name             = var.rName
  approved_patches = ["KB123456"]
}
