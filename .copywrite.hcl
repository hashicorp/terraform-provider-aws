# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

schema_version = 1

project {
  license        = "MPL-2.0"
  copyright_year = 2025
  copyright_holder = "IBM Corp."

  # (OPTIONAL) A list of globs that should not have copyright/license headers.
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  header_ignore = [
    ".ci/**",
    ".github/**",
    ".teamcity/**",
    ".release/**",
    "infrastructure/repository/labels-service.tf",
    ".goreleaser.yml",
  ]
}
