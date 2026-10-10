data "aws_bedrockagentcore_consent_portal" "test" {
{{- template "region" }}
  consent_portal_identifier = aws_bedrockagentcore_consent_portal.test.consent_portal_id
}
