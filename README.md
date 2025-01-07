<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform AWS Provider

[![Forums][discuss-badge]][discuss]

[discuss-badge]: https://img.shields.io/badge/discuss-terraform--aws-623CE4.svg?style=flat
[discuss]: https://discuss.hashicorp.com/c/terraform-providers/tf-aws/

The [AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) enables [Terraform](https://terraform.io) to manage [AWS](https://aws.amazon.com) resources.

- [Contributing guide](https://hashicorp.github.io/terraform-provider-aws/)
- [Quarterly development roadmap](ROADMAP.md)
- [FAQ](https://hashicorp.github.io/terraform-provider-aws/faq/)
- [Tutorials](https://learn.hashicorp.com/collections/terraform/aws-get-started)
- [discuss.hashicorp.com](https://discuss.hashicorp.com/c/terraform-providers/tf-aws/)
- [Google Groups](http://groups.google.com/group/terraform-tool)

_**Please note:** We take Terraform's security and our users' trust very seriously. If you believe you have found a security issue in the Terraform AWS Provider, please responsibly disclose it by contacting us at security@hashicorp.com._


This is the output for rc.Change.After, help me with tfjsonpath.New, example any attribute

map[bucket:tf-acc-test-7440854536379160617 error_document:[map[key:error.html]] expected_bucket_owner: id:tf-acc-test-7440854536379160617 index_document:[map[suffix:index.html]] redirect_all_requests_to:[] routing_rule:[map[condition:[map[http_error_code_returned_equals: key_prefix_equals:docs/]] redirect:[map[host_name: http_redirect_code: protocol: replace_key_prefix_with:documents/ replace_key_with:]]]] routing_rules:[{"Condition":{"KeyPrefixEquals":"docs/"},"Redirect":{"ReplaceKeyPrefixWith":"documents/"}}] website_domain:s3-website-us-west-2.amazonaws.com website_endpoint:tf-acc-test-7440854536379160617.s3-website-us-west-2.amazonaws.com]

map[acceleration_status: acl:<nil> arn:arn:aws:s3:::tf-acc-test-661749271477591619 bucket:tf-acc-test-661749271477591619 bucket_domain_name:tf-acc-test-661749271477591619.s3.amazonaws.com bucket_prefix: bucket_regional_domain_name:tf-acc-test-661749271477591619.s3.us-west-2.amazonaws.com cors_rule:[] force_destroy:false grant:[map[id:9af38a9971e9f66c76506a0b74aeaf25a9d6ca5a4648437573089ce314f27425 permissions:[FULL_CONTROL] type:CanonicalUser uri:]] hosted_zone_id:Z3BJ6K6RIION7M id:tf-acc-test-661749271477591619 lifecycle_rule:[] logging:[] object_lock_configuration:[] object_lock_enabled:false policy: region:us-west-2 replication_configuration:[] request_payer:BucketOwner server_side_encryption_configuration:[map[rule:[map[apply_server_side_encryption_by_default:[map[kms_master_key_id: sse_algorithm:AES256]] bucket_key_enabled:false]]]] tags:map[] tags_all:map[] timeouts:<nil> versioning:[map[enabled:false mfa_delete:false]] website:[map[error_document:error.html index_document:index.html redirect_all_requests_to: routing_rules:[{"Condition":{"KeyPrefixEquals":"docs/"},"Redirect":{"ReplaceKeyPrefixWith":"documents/"}}]]] website_domain:s3-website-us-west-2.amazonaws.com website_endpoint:tf-acc-test-661749271477591619.s3-website-us-west-2.amazonaws.com]