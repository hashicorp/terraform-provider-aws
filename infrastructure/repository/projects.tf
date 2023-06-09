// This configuration file sets up Actions variables for use when automating GitHub Projects.
// Currently, this consists only of the main team project proof of concept, but more may be added at a later date.

// Main Team Project
// -----------------

// Project ID (vars.team_project in workflow files)
resource "github_actions_variable" "team_project_id" {
  repository    = "terraform-provider-aws"
  variable_name = "team_project"
  value         = "PVT_kwDOAAuecM4AF-7h"
}

// Project's "Status" field's ID (vars.team_project_field_status in workflow files)
resource "github_actions_variable" "team_project_field_status_id" {
  repository    = "terraform-provider-aws"
  variable_name = "team_project_field_status"
  value         = "PVTSSF_lADOAAuecM4AF-7hzgDcsQA"
}

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
    "Backlog"       = "ef47b7a3",
    "Done"          = "98236657"
  }
}

resource "github_actions_variable" "team_project_field_status_option_ids" {
  for_each      = var.team_project_field_status_values
  repository    = "terraform-provider-aws"
  variable_name = "team_project_status_${replace(lower(each.key), " ", "_")}"
  value         = each.value
}
