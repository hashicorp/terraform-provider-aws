// A list of maintainers to be used as an "allow list" for various GitHub Actions.
// This allows us to make various "exceptions" for maintainers, such as automatically
// removing the `needs-triage` label from new Issues and Pull Requests
//
resource "github_actions_secret" "maintainer_list" {
  repository      = "terraform-provider-aws"
  secret_name     = "MAINTAINER_LIST"
  plaintext_value = "['breathingdust', 'dependabot[bot]', 'ewbankkit', 'gdavison', 'jar-b', 'johnsonaj', 'justinretzolk', 'marcosentino', 'YakDriver']"
}
