## 1.0.1 (November 09, 2017)

BUG FIXES:

* `tls_cert_request` and `tls_self_signed_cert` no longer cause a crash when `subject` isn't specified. ([#7](https://github.com/terraform-providers/terraform-provider-tls/issues/7))
* `tls_cert_request` and `tls_self_signed_cert` no longer generate empty-string values for various subject fields when they are not set in configuration. ([#10](https://github.com/terraform-providers/terraform-provider-tls/issues/10))

## 1.0.0 (September 15, 2017)

* No changes from 0.1.0; just adjusting to [the new version numbering scheme](https://www.hashicorp.com/blog/hashicorp-terraform-provider-versioning/).

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
