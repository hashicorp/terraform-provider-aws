resource "aws_ssm_patch_baseline" "test" {
  name                              = var.rName
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"

{{- template "tags" . }}
}
