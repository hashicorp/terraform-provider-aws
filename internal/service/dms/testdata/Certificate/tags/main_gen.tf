# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_certificate" "test" {
  certificate_id  = var.rName
  certificate_pem = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
