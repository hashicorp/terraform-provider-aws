<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Autonomous Database with a Write-Only Password

This example requires Terraform 1.11 or later because it configures the ADMIN
password with `admin_password_wo`. Run it from this directory with an existing
ODB network ID and the desired password supplied through the `odb_network_id`
and `autonomous_database_admin_password` input variables.

The other Oracle Database@AWS examples remain in the parent directory and are
validated separately with the provider's legacy Terraform versions.
