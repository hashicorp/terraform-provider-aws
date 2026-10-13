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

SMOKE_TESTS_LOGS=(
    TestAccLogsLogGroup_basic
    TestAccLogsLogGroup_multiple
)

SMOKE_TESTS_EC2=(
    TestAccVPCSecurityGroup_basic
    TestAccVPCSecurityGroup_egressMode
    TestAccVPCSecurityGroup_vpcAllEgress
    TestAccVPCSecurityGroupRule_race
    TestAccVPCSecurityGroupRule_protocolChange
    TestAccVPCDataSource_basic
    TestAccVPCSubnet_basic
    TestAccVPC_tenancy
    TestAccVPCRouteTableAssociation_Subnet_basic
    TestAccVPCRouteTable_basic
)

SMOKE_TESTS_ECS=(
    TestAccECSTaskDefinition_basic
    TestAccECSService_basic
)

SMOKE_TESTS_ELBV2=(
    TestAccELBV2TargetGroup_basic
)

SMOKE_TESTS_EVENTS=(
    TestAccEventsPutEventsAction_basic
)

SMOKE_TESTS_KMS=(
    TestAccKMSKey_basic
)

SMOKE_TESTS_LAMBDA=(
    TestAccLambdaFunction_basic
    TestAccLambdaPermission_basic
    TestAccLambdaCapacityProvider_List_basic
)

SMOKE_TESTS_META=(
    TestAccMetaRegionDataSource_basic
    TestAccMetaRegionDataSource_endpoint
    TestAccMetaPartitionDataSource_basic
)

SMOKE_TESTS_ROUTE53=(
    TestAccRoute53Record_basic_FullName
    TestAccRoute53Record_basic_ShortName
    TestAccRoute53Record_Latency_basic
    TestAccRoute53ZoneDataSource_name
)

SMOKE_TESTS_S3=(
    TestAccS3Bucket_Basic_basic
    TestAccS3Bucket_Security_corsUpdate
    TestAccS3BucketPublicAccessBlock_basic
    TestAccS3BucketPolicy_basic
    TestAccS3BucketACL_updateACL
    TestAccS3Object_basic
)

SMOKE_TESTS_SSM=(
    TestAccSSMParameterEphemeral_basic
)

SMOKE_TESTS_SECRETSMANAGER=(
    TestAccSecretsManagerSecret_basic
)

SMOKE_TESTS_STS=(
    TestAccSTSCallerIdentityDataSource_basic
)

SMOKE_TESTS_FUNCTION=(
    TestARNParseFunction_known
)

SMOKE_TESTS_REMAINING=(
    "${SMOKE_TESTS_LOGS[@]}"
    "${SMOKE_TESTS_EC2[@]}"
    "${SMOKE_TESTS_ECS[@]}"
    "${SMOKE_TESTS_ELBV2[@]}"
    "${SMOKE_TESTS_EVENTS[@]}"
    "${SMOKE_TESTS_KMS[@]}"
    "${SMOKE_TESTS_LAMBDA[@]}"
    "${SMOKE_TESTS_META[@]}"
    "${SMOKE_TESTS_ROUTE53[@]}"
    "${SMOKE_TESTS_S3[@]}"
    "${SMOKE_TESTS_SSM[@]}"
    "${SMOKE_TESTS_SECRETSMANAGER[@]}"
    "${SMOKE_TESTS_STS[@]}"
    "${SMOKE_TESTS_FUNCTION[@]}"
)

parallelism_args=()
if [[ -n "${PACKAGE_PARALLELISM:-}" ]]; then
    parallelism_args=(-p "${PACKAGE_PARALLELISM}")
fi

# IAM is foundational to most other services, so run its tests first

printf -v run_iam '%s|' "${SMOKE_TESTS_IAM[@]}"

TF_ACC=1 "${GO_BIN}" test \
    ./internal/service/iam/... \
    "${parallelism_args[@]}" -count 1 -timeout 60m -vet=off -buildvcs=false \
    -run="${run_iam%|}"

printf -v run_remaining '%s|' "${SMOKE_TESTS_REMAINING[@]}"

TF_ACC=1 "${GO_BIN}" test \
    ./internal/service/logs/... \
    ./internal/service/ec2/... \
    ./internal/service/ecs/... \
    ./internal/service/elbv2/... \
    ./internal/service/events/... \
    ./internal/service/kms/... \
    ./internal/service/lambda/... \
    ./internal/service/meta/... \
    ./internal/service/route53/... \
    ./internal/service/s3/... \
    ./internal/service/ssm/... \
    ./internal/service/secretsmanager/... \
    ./internal/service/sts/... \
    ./internal/function/... \
    "${parallelism_args[@]}" -count 1 -timeout 60m -vet=off -buildvcs=false \
    -run="${run_remaining%|}"
