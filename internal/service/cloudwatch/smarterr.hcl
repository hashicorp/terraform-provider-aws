# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

parameter "service" {
  value = "CloudWatch"
}

hint "dashboard_name_conflict" {
  error_contains = "DashboardAlreadyExists"
  suggestion = "A dashboard with this name already exists in your AWS account. Choose a unique name for your CloudWatch dashboard, or import the existing dashboard into Terraform using `terraform import` if you want to manage it."
}
