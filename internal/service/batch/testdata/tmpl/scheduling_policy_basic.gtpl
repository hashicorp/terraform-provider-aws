resource "aws_batch_scheduling_policy" "test" {
  name = var.rName

  fair_share_policy {
    compute_reservation = 0
    share_decay_seconds = 0
  }
{{- template "tags" . }}
}
