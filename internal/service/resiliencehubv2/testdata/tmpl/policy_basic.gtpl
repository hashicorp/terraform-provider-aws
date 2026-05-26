resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  availability_slo {
    target = 99.9
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }

{{- template "tags" . }}
}
