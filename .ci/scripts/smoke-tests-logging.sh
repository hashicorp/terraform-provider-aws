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

# Response contains long lines
SMOKE_TESTS_S3=(
    TestAccS3ObjectsDataSource_maxKeysLarge
)

# Response contains long lines
SMOKE_TESTS_SSM=(
    TestAccSSMResourceDataSync_basic
    TestAccSSMResourceDataSync_Update_s3DestinationPrefix
)

SMOKE_TESTS=(
    "${SMOKE_TESTS_IAM[@]}"
    "${SMOKE_TESTS_S3[@]}"
    "${SMOKE_TESTS_SSM[@]}"
)

printf -v run_tests '%s|' "${SMOKE_TESTS[@]}"

TF_ACC=1 "${GO_BIN}" test \
    ./internal/service/iam/... \
    ./internal/service/s3/... \
    ./internal/service/ssm/... \
    -count 1 -timeout 60m -vet=off -buildvcs=false \
    -run="${run_tests%|}"
