#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
    conf=$(pwd)/aws.conf

    function cleanup {
        rm "${conf}"
    }
    trap cleanup EXIT

    touch "${conf}"
    chmod 600 "${conf}"
    cat <<EOF >"${conf}"
[profile perftest]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=perftest
fi

unset TF_LOG

function tester {
    local service=$1
    local tests=$2

    local results
    results=$(TF_ACC=1 go test ./internal/service/"${service}"/... -v -parallel 4 -run="${tests}" -timeout 60m -count 1 -vet=off -buildvcs=false  2>&1)
    local exit_code=$?

    echo "${results}"

    if [[ "${results}" == *"text file busy"* ]]; then
        echo "FAILED attempt to run tests"
        echo "Trying again..."
        sleep 5
        tester "${service}" "${tests}"
    fi

    if [[ "${exit_code}" -ne 0 ]]; then
        exit "${exit_code}"
    fi
}

if [[ ! -f "iamsanity.test" ]]; then
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
    printf -v iam_tests '^%s$|' "${SMOKE_TESTS_IAM[@]}"
    tester "iam" "${iam_tests%|}"
    touch iamsanity.test
    exit 0
fi

if [[ ! -f "logssanity.test" ]]; then
    SMOKE_TESTS_LOGS=(
        TestAccLogsLogGroup_basic
        TestAccLogsLogGroup_multiple
    )
    printf -v logs_tests '^%s$|' "${SMOKE_TESTS_LOGS[@]}"
    tester "logs" "${logs_tests%|}"
    touch logssanity.test
    exit 0
fi

if [[ ! -f "ec2sanity.test" ]]; then
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
    printf -v ec2_tests '^%s$|' "${SMOKE_TESTS_EC2[@]}"
    tester "ec2" "${ec2_tests%|}"
    touch ec2sanity.test
    exit 0
fi

if [[ ! -f "ecssanity.test" ]]; then
    SMOKE_TESTS_ECS=(
        TestAccECSTaskDefinition_basic
        TestAccECSService_basic
    )
    printf -v ecs_tests '^%s$|' "${SMOKE_TESTS_ECS[@]}"
    tester "ecs" "${ecs_tests%|}"
    touch ecssanity.test
    exit 0
fi

if [[ ! -f "elbv2sanity.test" ]]; then
    SMOKE_TESTS_ELBV2=(
        TestAccELBV2TargetGroup_basic
    )
    printf -v elbv2_tests '^%s$|' "${SMOKE_TESTS_ELBV2[@]}"
    tester "elbv2" "${elbv2_tests%|}"
    touch elbv2sanity.test
    exit 0
fi

if [[ ! -f "eventssanity.test" ]]; then
    SMOKE_TESTS_EVENTS=(
        TestAccEventsPutEventsAction_basic
    )
    printf -v events_tests '^%s$|' "${SMOKE_TESTS_EVENTS[@]}"
    tester "events" "${events_tests%|}"
    touch eventssanity.test
    exit 0
fi

if [[ ! -f "kmssanity.test" ]]; then
    tester "kms" 'TestAccKMSKey_basic'
    touch kmssanity.test
    exit 0
fi

if [[ ! -f "lambdasanity.test" ]]; then
    tester "lambda" 'TestAccLambdaFunction_basic|TestAccLambdaPermission_basic'
    touch lambdasanity.test
    exit 0
fi

if [[ ! -f "metasanity.test" ]]; then
    tester "meta" 'TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic'
    touch metasanity.test
    exit 0
fi

if [[ ! -f "route53sanity.test" ]]; then
    tester "route53" 'TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name'
    touch route53sanity.test
    exit 0
fi

if [[ ! -f "s3sanity.test" ]]; then
    tester "s3" 'TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL'
    touch s3sanity.test
    exit 0
fi

if [[ ! -f "secretsmanagersanity.test" ]]; then
    tester "secretsmanager" 'TestAccSecretsManagerSecret_basic'
    touch secretsmanagersanity.test
    exit 0
fi

if [[ ! -f "stssanity.test" ]]; then
    tester "sts" 'TestAccSTSCallerIdentityDataSource_basic'
    touch stssanity.test
    exit 0
fi

echo "##teamcity[notification notifier='slack' message='*Sanity Tests Passed!*:white_check_mark:' sendTo='CN0G9S7M4' connectionId='PROJECT_EXT_8']"
