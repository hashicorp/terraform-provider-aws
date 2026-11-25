# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Reference: https://github.com/hashicorp/security-scanner/blob/main/CONFIG.md#binary (private repository)

binary {
  secrets {
    all = true
  }
  go_modules   = true
  osv          = true
  oss_index    = false
  nvd          = false

  triage {
    suppress {
      vulnerabilities = [
        // golang.org/x/crypto/openpgp is deprecated/unmaintained with no fixed
        // version. The provider does not use this package (confirmed via
        // `go mod why golang.org/x/crypto/openpgp`); it imports the maintained
        // fork github.com/ProtonMail/go-crypto/openpgp instead. x/crypto is
        // only pulled in for the unaffected golang.org/x/crypto/cryptobyte
        // package. Not reachable in the built binary.
        "GO-2026-5932",
      ]
    }
  }
}