resource "aws_ivs_playback_key_pair" "test" {
{{- template "region" }}
  public_key = var.rTlsEcdsaPublicKeyPem
{{- template "tags" }}
}
