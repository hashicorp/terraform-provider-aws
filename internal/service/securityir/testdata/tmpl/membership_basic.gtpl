resource "aws_securityir_membership" "test" {
{{- template "region" }}
  membership_name = var.rName

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = "${var.rName}-1@example.com"
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = "${var.rName}-2@example.com"
  }
{{- template "tags" . }}
}
