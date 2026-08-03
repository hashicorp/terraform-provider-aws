resource "aws_resiliencehubv2_system" "test" {
{{- template "region" }}
  name = "${var.rName}-system"
}

resource "aws_resiliencehubv2_user_journey" "test" {
{{- template "region" }}
  name       = var.rName
  system_arn = aws_resiliencehubv2_system.test.arn
}
