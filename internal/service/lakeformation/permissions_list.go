// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// The List Resource `aws_lakeformation_permissions` will not be implemented: This is a "Composite Grant Resource".
// The AWS Lake Formation API does not provide stable per-grant identifiers, and returned grants are normalized in
// shapes that depend on the original Terraform input configuration. As a consequence, this resource is not
// importable and does not have a Resource Identity schema.
//
// See docs/design-decisions/resource-types-without-list.md
