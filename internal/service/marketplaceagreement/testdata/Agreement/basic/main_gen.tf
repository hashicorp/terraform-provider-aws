# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# New Relic Kubernetes Operator: a free public offer whose terms are a LegalTerm and a SupportTerm.
resource "aws_marketplaceagreement_agreement" "test" {
  agreement_proposal_id = "at-32xgdgxyhupql1jw55dnk7ff1"

  requested_term {
    id = "term-7e5d56619c3a08143428c3771d7f6de4bf687508be034e6c4b4dd26c1c1d966e"
  }

  requested_term {
    id = "term-f5bdd147164b3f3b593e9b56a886c9e0fa3cd4fd9e7cf832ab13ef88161c14c8"
  }
}

