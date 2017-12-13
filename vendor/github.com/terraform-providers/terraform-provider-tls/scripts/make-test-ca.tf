// running 'terraform apply' will create 'testCAPrivateKey.pem' and 'testCACert.pem'
// copy their contents into the corresponding constants in 'tls/provider_test.go'

resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits = 1024
}

resource "local_file" "ca_key" {
  filename = "testCAPrivateKey.pem"
  content = "${tls_private_key.ca.private_key_pem}"
}

resource "tls_self_signed_cert" "ca" {
  is_ca_certificate = true
  key_algorithm = "RSA"
  allowed_uses = [
    "cert_signing"
  ]
  private_key_pem = "${tls_private_key.ca.private_key_pem}"
  "subject" {
    organization = "Example, Inc"
    organizational_unit = "Department of CA Testing"
    common_name = "root"
    country = "US"
    province = "CA"
    locality = "Pirate Harbor"
  }
  validity_period_hours = 87600
}

resource "local_file" "ca_cert" {
  filename = "testCACert.pem"
  content = "${tls_self_signed_cert.ca.cert_pem}"
}
