# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider_source = "registry.terraform.io/hashicorp/aws"
provider_dir    = "."
schema_json     = "terraform-providers-schema/schema.json"

file_aliases = {
  "list_resource/aws_ebs_volume" = "aws_ec2_ebs_volume"
}

ignore_contents_check = [
  "data_source/aws_kms_secret",
]

# ─── Type overrides ─────────────────────────────────────────────────────────
# Define bylines, frontmatter requirements, and other AWS-specific conventions.

type "resource" {
  schema_kind   = "resource"
  website_paths = ["website/docs/r/{name}.html.markdown"]
  title_prefix  = "Resource"

  arguments_bylines = [
    "This resource supports the following arguments:",
    "The following arguments are required:",
    "The following arguments are optional:",
    "This resource does not support any arguments.",
  ]
  attributes_bylines = [
    "This resource exports the following attributes in addition to the arguments above:",
    "This resource exports no additional attributes.",
  ]

  require_attributes = "required"
  require_import     = "optional"
  require_timeouts   = "optional"
  require_signature  = "forbidden"

  frontmatter_require = ["description", "page_title"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = true
}

type "data_source" {
  schema_kind   = "data_source"
  website_paths = ["website/docs/d/{name}.html.markdown"]
  title_prefix  = "Data Source"

  arguments_bylines = [
    "This data source supports the following arguments:",
    "The following arguments are required:",
    "The following arguments are optional:",
    "This data source does not support any arguments.",
  ]
  attributes_bylines = [
    "This data source exports the following attributes in addition to the arguments above:",
    "This data source exports no additional attributes.",
  ]

  require_attributes = "required"
  require_import     = "forbidden"
  require_timeouts   = "optional"
  require_signature  = "forbidden"

  frontmatter_require = ["description", "page_title"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = true
}

type "ephemeral" {
  schema_kind   = "ephemeral"
  website_paths = ["website/docs/ephemeral-resources/{name}.html.markdown"]
  title_prefix  = "Ephemeral"

  arguments_bylines = [
    "This ephemeral resource supports the following arguments:",
    "The following arguments are required:",
    "The following arguments are optional:",
    "This ephemeral resource does not support any arguments.",
  ]
  attributes_bylines = [
    "This ephemeral resource exports the following attributes in addition to the arguments above:",
    "This ephemeral resource exports no additional attributes.",
  ]

  require_attributes = "required"
  require_import     = "forbidden"
  require_timeouts   = "forbidden"
  require_signature  = "forbidden"

  frontmatter_require = ["description", "page_title"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = true
}

type "function" {
  schema_kind   = "function"
  website_paths = ["website/docs/functions/{name}.html.markdown"]
  title_prefix  = "Function"

  arguments_heading              = "Arguments"
  allow_missing_arguments_byline = true

  require_attributes = "forbidden"
  require_import     = "forbidden"
  require_timeouts   = "forbidden"
  require_signature  = "required"

  frontmatter_require = ["description", "page_title"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = false
}

type "list_resource" {
  schema_kind   = "list_resource"
  website_paths = ["website/docs/list-resources/{name}.html.markdown"]
  title_prefix  = "List Resource"

  arguments_bylines = [
    "This list resource supports the following arguments:",
    "The following arguments are required:",
    "The following arguments are optional:",
    "This list resource does not support any arguments.",
  ]

  require_attributes = "forbidden"
  require_import     = "forbidden"
  require_timeouts   = "forbidden"
  require_signature  = "forbidden"

  frontmatter_require = ["description", "page_title"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = true
}

type "action" {
  schema_kind   = "action"
  website_paths = ["website/docs/actions/{name}.html.markdown"]
  title_prefix  = "Action"

  arguments_bylines = [
    "This action supports the following arguments:",
    "The following arguments are required:",
    "The following arguments are optional:",
    "This action does not support any arguments.",
  ]

  require_attributes = "forbidden"
  require_import     = "forbidden"
  require_timeouts   = "forbidden"
  require_signature  = "forbidden"

  frontmatter_require = ["description", "page_title", "subcategory"]
  frontmatter_forbid  = ["sidebar_current"]

  region_aware = false
}

# ─── Check blocks ───────────────────────────────────────────────────────────

check "schema_docs" {
  enabled = true

  byline      = true
  coverage    = true
  description = true
  format      = true
  heading     = true
  labels      = true
  ordering    = true

  block_heading_styles = [
    "`{Parent}` `{Block}` Block",
    "`{Block}` Block",
    "{Block} Block",
    "{Block} block",
    "{Block} Configuration Block",
    "{Block} Argument Reference",
    "{Block} Attribute Reference",
    "{Title} Arguments",
    "{Title} Argument Reference",
    "{Title} Attribute Reference",
    "`{Block}`",
    "{Block}",
    "{Title}",
  ]

  prefer_block_heading_styles = [
    "`{Parent}` `{Block}` Block",
    "`{Block}` Block",
  ]
}

check "import_section" {
  enabled = true
  require_identity_section = true
}

check "frontmatter" {
  enabled = true

  require_subcategory = true
  require_page_title  = true
  require_description = true
  require_layout      = true

  allow_subcategories_file = "website/allowed-subcategories.txt"

  allow_empty_subcategory_targets = [
    "arn_build",
    "arn_parse",
    "trim_iam_role_path",
    "user_agent",
  ]
}

check "section_presence" {
  enabled = true
}

check "timeouts_section" {
  enabled = true
}

check "region_argument" {
  enabled = true
}

check "file_check" {
  enabled = true
  max_file_size = 500000
  allow_extensions = [".html.markdown"]
  allow_registry_extensions = [".md"]
  inline_links = true
}

check "file_match" {
  enabled = true
  ignore_missing = [
    "aws_alb",
    "aws_alb_listener",
    "aws_alb_listener_certificate",
    "aws_alb_listener_rule",
    "aws_alb_target_group",
    "aws_alb_target_group_attachment",
    "aws_alb_trust_store",
    "aws_alb_trust_store_revocation",
    "aws_albs",
  ]
}
