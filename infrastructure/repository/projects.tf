# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

// This configuration file sets up Actions variables for use when automating GitHub Projects.
// Currently, this consists only of the main team project proof of concept, but more may be added at a later date.


// Project's "Status" field's options IDs (vars.team_project_status_${option_name_snake_case}
// This set of variables will take the names of each possible value for the "Status" column, convert them to snake case
// and prefix them with "team_project_status_". E.g. "To Do" becomes "team_project_status_to_do".
variable "team_project_field_status_values" {
  type        = map(string)
  description = "A mapping of the statuses in the main team project to their IDs"
  default = {
    "To Do"         = "f75ad846",
    "In Progress"   = "47fc9ee4",
    "Waiting"       = "e85f2e5d",
    "Maintainer PR" = "28a034bc",
    "Pending Merge" = "043bc06e",
    "Done"          = "98236657"
  }
}

variable "team_project_field_view_values" {
  type        = map(string)
  description = "A mapping of the options in the view field in the main team project to their IDs"
  default = {
    "Working Board"          = "8d366764",
    "Partner Contribution"   = "b0492564",
    "External Maintainer"    = "f17c472c",
    "S3 Backend"             = "5c656f1e",
    "Engineering Initiative" = "a62d09b9"
  }
}

resource "github_actions_variable" "team_project_field_status_option_ids" {
  for_each      = var.team_project_field_status_values
  repository    = "terraform-provider-aws"
  variable_name = "team_project_status_${replace(lower(each.key), " ", "_")}"
  value         = each.value
}

resource "github_actions_variable" "team_project_field_view_option_ids" {
  for_each      = var.team_project_field_view_values
  repository    = "terraform-provider-aws"
  variable_name = "team_project_view_${replace(lower(each.key), " ", "_")}"
  value         = each.value
}
