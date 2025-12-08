# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

// Reference: https://github.com/hashicorp/crt-core-helloworld/blob/main/.release/ci.hcl (private repository)
//
// One way to validate this file, with a local build of the orchestrator (an internal repo):
//
// $ GITHUB_TOKEN="not-used" orchestrator parse config -use-v2 -local-config=.release/ci.hcl

schema = "2"

project "terraform-provider-aws" {
  // team is currently unused and has no meaning
  // but is required to be non-empty by CRT orchestator
  team = "_UNUSED_"

  slack {
    notification_channel = "C09L3G3V55J" // #feed-tf-aws
  }

  github {
    organization     = "hashicorp"
    repository       = "terraform-provider-aws"
    release_branches = ["main", "release/**"]
  }
}

event "merge" {
}

event "build" {
  action "build" {
    depends = ["merge"]

    organization = "hashicorp"
    repository   = "terraform-provider-aws"
    workflow     = "build"
  }
}

event "prepare" {
  # `prepare` is the Common Release Tooling (CRT) artifact processing workflow.
  # It prepares artifacts for potential promotion to staging and production.
  # For example, it scans and signs artifacts.

  depends = ["build"]

  action "prepare" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "prepare"
    depends      = ["build"]
  }

  notification {
    on = "fail"
  }
}

event "trigger-staging" {
}

event "promote-staging" {
  action "promote-staging" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-staging"
    depends      = null
    config       = "release-metadata.hcl"
  }

  depends = ["trigger-staging"]

  notification {
    on = "always"
  }
}

event "trigger-production" {
}

event "promote-production" {
  action "promote-production" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-production"
    depends      = null
    config       = ""
  }

  promotion-events {
    bump-version-patch = true

    post-promotion {
      organization = "hashicorp"
      repository   = "terraform-provider-aws"
      workflow     = "crt-post-promote-production.yml"
    }
  }

  depends = ["trigger-production"]

  notification {
    on = "always"
  }
}

