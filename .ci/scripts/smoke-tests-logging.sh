#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

GO_BIN="${GO_BIN:-go}"

SMOKE_TESTS_IAM=(
    TestAccIAMRole_basic
    TestAccIAMRole_namePrefix
    TestAccIAMRole_disappears
    TestAccIAMRole_InlinePolicy_basic
    TestAccIAMPolicyDocumentDataSource_basic
    TestAccIAMPolicyDocumentDataSource_sourceConflicting
    TestAccIAMPolicyDocumentDataSource_sourcePolicyValidJSON
    TestAccIAMRolePolicyAttachment_basic
    TestAccIAMRolePolicyAttachment_disappears
    TestAccIAMRolePolicyAttachment_Disappears_role
    TestAccIAMPolicy_basic
    TestAccIAMPolicy_policy
    TestAccIAMPolicy_tags
    TestAccIAMRolePolicy_basic
    TestAccIAMRolePolicy_unknownsInPolicy
    TestAccIAMInstanceProfile_basic
    TestAccIAMInstanceProfile_tags
    TestAccIAMPolicy_List_basic
    TestAccIAMRole_Identity_basic
)

printf -v run_iam '%s|' "${SMOKE_TESTS_IAM[@]}"

TF_ACC=1 "${GO_BIN}" test \
    ./internal/service/iam/... \
    -count 1 -timeout 60m -vet=off -buildvcs=false \
    -run="${run_iam%|}"
