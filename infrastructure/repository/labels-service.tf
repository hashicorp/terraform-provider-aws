data "github_branch" "main" {
  branch     = "main"
  repository = "terraform-provider-aws"
}

data "github_tree" "main" {
  recursive  = true
  repository = data.github_branch.main.repository
  tree_sha   = data.github_branch.main.sha
}

locals {

  # A breakdown on this becuase while I realize it's a pretty ugly one-liner,
  # I wasn't able to come up with a prettier way to do it, outside of just
  # breaking the `compact()` and `toset()` functions out into separate locals.
  #
  # The `for` bit in here takes the list of tree entries from the data source
  # uses `regexall` to check to see if the path for the entry is a directory
  # under the `internal/service/` directory. `regexall()` returns a list of
  # every match, so by testing whether the length is greater than 0, we can
  # determine if the filepath matches our regex.
  #
  # If the path matches the regex, we then trim the `internal/service/` prefix
  # before adding the path to our resulting list. If the path does not match,
  # we pass `null` instead.
  #
  # As a note, it would probably be more efficient to *not* trim 'service/`,
  # however, we would need to do some state surgery to move over to that, since
  # we were previously supplying a static list of service names rather than
  # using these data sources.
  #
  # Passing `null` means that we can wrap the whole list to `compact()` to
  # remove all of the `null` values and leave only the list of tag values.
  # Further, passing this list to `toset()` means that we have something that
  # can be used by `for_each` in the label resource below.

  services = toset(compact([
    for entry in data.github_tree.main.entries :
    length(regexall("^internal\\/service\\/[[:alnum:]]+$", entry.path)) > 0 ? trimprefix(entry.path, "internal/service/") : null
  ]))

}


resource "github_issue_label" "service" {
  for_each = local.services

  repository  = "terraform-provider-aws"
  name        = "service/${each.value}"
  color       = "7b42bc" # color:terraform (logomark)
  description = "Issues and PRs that pertain to the ${each.value} service."
}
