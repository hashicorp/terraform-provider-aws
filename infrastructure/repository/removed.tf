# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# TODO: delete after removal
# https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources

removed {
  from = github_actions_variable.team_project_field_status_option_ids

  lifecycle {
    destroy = false
  }
}

removed {
  from = github_actions_variable.team_project_field_view_option_ids

  lifecycle {
    destroy = false
  }
}

removed {
  from = github_actions_secret.core_contributors

  lifecycle {
    destroy = false
  }
}

removed {
  from = github_actions_secret.maintainers

  lifecycle {
    destroy = false
  }
}

removed {
  from = github_actions_secret.partners

  lifecycle {
    destroy = false
  }
}
